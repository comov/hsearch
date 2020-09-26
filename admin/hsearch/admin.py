from django.contrib import admin
from django.contrib.admin.sites import site as default_site
from django.contrib.auth.admin import UserAdmin, GroupAdmin
from django.contrib.auth.models import Group, User
from django.db import models
from django.utils.safestring import SafeString
from hsearch.admin_inlines import FeedbackInline, AnswerInline, ImageInline
from hsearch.forms import AdminAuthenticationForm
from hsearch.models import Chat, Offer, Answer, Feedback, Image, TgMessage


def _yes_no_img(var):
    res = ('yes', 'True') if var else ('no', 'False')
    return '<img src="/static/admin/img/icon-%s.svg" alt="%s">' % res


class AdminSite(admin.AdminSite):
    login_form = AdminAuthenticationForm
    login_template = 'admin/login.html'

    def _registry_getter(self):
        return default_site._registry

    def _registry_setter(self, value):
        default_site._registry = value

    _registry = property(_registry_getter, _registry_setter)


site = AdminSite()
admin.site = site
default_site.enable_nav_sidebar = False

admin.site.register(Group, GroupAdmin)
admin.site.register(User, UserAdmin)


@admin.register(Chat)
class ChatAdmin(admin.ModelAdmin):
    list_display = [
        'display',
        'telegram_link',
        'c_type',
        'sites',
        'other_filters',
        'enable',
        'created',
    ]

    list_filter = [
        'c_type',
        'enable',
        'diesel',
        'lalafo',
        'house',
        'photo',
    ]

    search_fields = [
        'title',
        'username',
    ]

    inlines = [
        FeedbackInline,
        AnswerInline,
    ]

    ordering = [
        '-created',
    ]

    def display(self, obj: Chat):
        return f'{obj.title} (#{obj.id})'

    display.short_description = 'display'

    def telegram_link(self, obj: Chat):
        if not obj.username:
            return '-'
        return SafeString(f'<a href="https://t.me/{obj.username}">{obj.username}</a>')

    telegram_link.short_description = 'telegram'

    def sites(self, obj: Chat):
        return SafeString(
            f'diesel: {_yes_no_img(obj.diesel)}<br>'
            f'lalafo: {_yes_no_img(obj.lalafo)}<br>'
            f'house: {_yes_no_img(obj.house)}',
        )

    sites.short_description = 'sites'

    def other_filters(self, obj: Chat):
        return SafeString(
            f'usd: {obj.usd}<br>'
            f'kgs: {obj.kgs}<br>'
            f'photo: {_yes_no_img(obj.photo)}<br>'
        )

    other_filters.short_description = 'other filters'


@admin.register(Offer)
class OfferAdmin(admin.ModelAdmin):
    search_fields = [
        'topic',
        'body',
    ]

    list_display = [
        'topic',
        'site_link',
        'floor',
        'area',
        'room_numbers',
        'full_price',
        'images_count',
        'phone_count',
        'created',
    ]

    list_filter = [
        'site',
        'room_numbers',
        'currency',
        'floor',
    ]

    readonly_fields = [
        'images_count',
    ]

    inlines = [
        ImageInline,
    ]

    ordering = [
        '-created',
    ]

    phones_cache = {}

    def get_queryset(self, request):
        qs = super().get_queryset(request)
        res = qs.values('phone').annotate(models.Count('id')).order_by()
        self.phones_cache = {i['phone']: i['id__count'] for i in res}
        return qs

    def site_link(self, obj: Offer):
        return SafeString(f'<a href="{obj.url}" target="_blank">{obj.site.title()}</a>')

    site_link.short_description = 'site'

    def phone_count(self, obj: Offer):
        if obj.phone == '':
            return '-'
        phone_count = self.phones_cache.get(obj.phone) or 0
        if phone_count < 3:
            return SafeString(f'<a href="tel:{obj.phone}">{obj.phone}</a>')
        return SafeString(f'<a href="tel:{obj.phone}" style="color:red;">{obj.phone} ({phone_count})</a>')

    phone_count.short_description = 'phone'


@admin.register(Answer)
class AnswerAdmin(admin.ModelAdmin):
    list_display = [
        'id',
        'chat_link',
        'offer_link',
        'dislike',
        'created',
    ]

    ordering = [
        '-created',
    ]

    def chat_link(self, obj: Feedback):
        return SafeString(f'<a href="/admin/hsearch/chat/{obj.chat.id}/">{obj.chat}</a>')

    chat_link.short_description = 'chat'

    def offer_link(self, obj: Feedback):
        return SafeString(f'<a href="/admin/hsearch/offer/{obj.offer.id}/">{obj.offer}</a>')

    offer_link.short_description = 'offer'


@admin.register(Feedback)
class FeedbackAdmin(admin.ModelAdmin):
    list_display = [
        'id',
        'chat_link',
        'telegram_link',
        'body',
        'created',
    ]

    search_fields = [
        'username',
        'chat__title',
        'body',
    ]

    ordering = [
        '-created',
    ]

    def telegram_link(self, obj: Feedback):
        if not obj.username:
            return '-'
        return SafeString(f'<a href="https://t.me/{obj.username}">{obj.username}</a>')

    telegram_link.short_description = 'telegram'

    def chat_link(self, obj: Feedback):
        return SafeString(f'<a href="/admin/hsearch/chat/{obj.chat.id}/">{obj.chat}</a>')

    chat_link.short_description = 'chat'


@admin.register(Image)
class ImageAdmin(admin.ModelAdmin):
    list_display = [
        'path',
        'offer_link',
        'image',
        'created',
    ]

    autocomplete_fields = [
        'offer',
    ]

    search_fields = [
        'offer__topic',
        'path',
    ]

    ordering = [
        '-created',
    ]

    def image(self, obj: Image):
        name = obj.path.split('/')[-1]
        return SafeString(f'<img height="200px" src="{obj.path}" alt="{name}"/>')

    image.short_description = 'image'

    def offer_link(self, obj: Feedback):
        return SafeString(f'<a href="/admin/hsearch/offer/{obj.offer.id}/">{obj.offer}</a>')

    offer_link.short_description = 'offer'


@admin.register(TgMessage)
class TgMessageAdmin(admin.ModelAdmin):
    list_display = [
        'message',
        'chat_link',
        'offer_link',
        'kind',
        'created',
    ]

    autocomplete_fields = [
        'offer',
        'chat',
    ]

    list_filter = [
        'kind',
    ]

    search_fields = [
        'chat__title',
        'offer__topic',
    ]

    ordering = [
        '-created',
    ]

    def chat_link(self, obj: Feedback):
        return SafeString(f'<a href="/admin/hsearch/chat/{obj.chat.id}/">{obj.chat}</a>')

    chat_link.short_description = 'chat'

    def offer_link(self, obj: Feedback):
        return SafeString(f'<a href="/admin/hsearch/offer/{obj.offer.id}/">{obj.offer}</a>')

    offer_link.short_description = 'offer'
