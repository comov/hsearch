from django.contrib import admin
from django.contrib.admin.sites import site as default_site
from django.contrib.auth.admin import GroupAdmin, UserAdmin
from django.contrib.auth.models import Group, User
from django.db import models
from django.utils.safestring import SafeString

from hsearch.admin_inlines import AnswerInline, FeedbackInline, ImageInline
from hsearch.forms import AdminAuthenticationForm
from hsearch.models import Apartment, Answer, Chat, Feedback, Image, TgMessage


def _yes_no_img(var):
    res = ("yes", "True") if var else ("no", "False")
    return '<img src="/static/admin/img/icon-%s.svg" alt="%s">' % res


class AdminSite(admin.AdminSite):
    login_form = AdminAuthenticationForm
    login_template = "admin/login.html"

    def _registry_getter(self):
        return default_site._registry

    def _registry_setter(self, value):
        default_site._registry = value

    _registry = property(_registry_getter, _registry_setter)


site = AdminSite()
site.enable_nav_sidebar = False
admin.site = site
default_site.enable_nav_sidebar = False

admin.site.register(Group, GroupAdmin)
admin.site.register(User, UserAdmin)


@admin.register(Chat)
class ChatAdmin(admin.ModelAdmin):
    list_display = [
        "display",
        "telegram_link",
        "c_type",
        "sites",
        "other_filters",
        "enable",
        "created",
    ]

    list_filter = [
        "c_type",
        "enable",
        "diesel",
        "lalafo",
        "house",
        "photo",
    ]

    search_fields = [
        "title",
        "username",
    ]

    inlines = [
        FeedbackInline,
        AnswerInline,
    ]

    ordering = [
        "-created",
    ]

    def display(self, obj: Chat):
        return f"{obj.title} (#{obj.id})"

    display.short_description = "display"

    def telegram_link(self, obj: Chat):
        if not obj.username:
            return "-"
        return SafeString(f'<a href="https://t.me/{obj.username}">{obj.username}</a>')

    telegram_link.short_description = "telegram"

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


@admin.register(Apartment)
class ApartmentAdmin(admin.ModelAdmin):
    search_fields = [
        "topic",
        "body",
    ]

    list_display = [
        "url",
        "topic",
        "phone",
        "rooms",
        "body",
        "images_count",
        "price",
        "currency",
        "area",
        "city",
        "room_type",
        "site",
        "floor",
        "max_floor",
        "district",
        "lat",
        "lon",
        "created",
    ]

    list_filter = [
        "site",
        "rooms",
        "currency",
        "floor",
    ]

    readonly_fields = [
        "images_count",
    ]

    inlines = [
        ImageInline,
    ]

    ordering = [
        "-created",
    ]

    phones_cache = {}

    def get_queryset(self, request):
        qs = super().get_queryset(request)
        res = qs.values('phone').annotate(models.Count('id')).order_by()
        self.phones_cache = {i['phone']: i['id__count'] for i in res}
        return qs

    def site_link(self, obj: Apartment):
        return SafeString(f'<a href="{obj.url}" target="_blank">{obj.site.title()}</a>')

    site_link.short_description = 'site'

    def phone_count(self, obj: Apartment):
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
        "id",
        "chat_link",
        "apartment_link",
        "dislike",
        "created",
    ]

    ordering = [
        "-created",
    ]

    def chat_link(self, obj: Answer):
        return SafeString(f'<a href="/hsearch/hsearch/chat/{obj.chat.id}/">{obj.chat}</a>')

    chat_link.short_description = 'chat'

    def apartment_link(self, obj: Answer):
        return SafeString(f'<a href="/hsearch/hsearch/apartment/{obj.apartment.id}/">{obj.apartment}</a>')

    apartment_link.short_description = 'Apartment'


@admin.register(Feedback)
class FeedbackAdmin(admin.ModelAdmin):
    list_display = [
        "id",
        "chat_link",
        "telegram_link",
        "body",
        "created",
    ]

    search_fields = [
        "username",
        "chat__title",
        "body",
    ]

    ordering = [
        "-created",
    ]

    def telegram_link(self, obj: Feedback):
        if not obj.username:
            return '-'
        return SafeString(f'<a href="https://t.me/{obj.username}">{obj.username}</a>')

    telegram_link.short_description = 'telegram'

    def chat_link(self, obj: Feedback):
        return SafeString(f'<a href="/hsearch/hsearch/chat/{obj.chat.id}/">{obj.chat}</a>')

    chat_link.short_description = 'chat'


@admin.register(Image)
class ImageAdmin(admin.ModelAdmin):
    list_display = [
        "path",
        "apartment_link",
        "image",
        "created",
    ]

    autocomplete_fields = [
        "apartment",
    ]

    search_fields = [
        "apartment__topic",
        "path",
    ]

    ordering = [
        "-created",
    ]

    def image(self, obj: Image):
        name = obj.path.split('/')[-1]
        return SafeString(f'<img height="200px" src="{obj.path}" alt="{name}"/>')

    image.short_description = "image"

    def apartment_link(self, obj: Image):
        return SafeString(f'<a href="/hsearch/hsearch/apartment/{obj.apartment.id}/">{obj.apartment}</a>')

    apartment_link.short_description = 'Apartment'


@admin.register(TgMessage)
class TgMessageAdmin(admin.ModelAdmin):
    list_display = [
        "message_id",
        "chat_link",
        "apartment_link",
        "kind",
        "created",
    ]

    autocomplete_fields = [
        "apartment",
        "chat",
    ]

    list_filter = [
        "kind",
    ]

    search_fields = [
        "chat__title",
        "apartment__topic",
    ]

    ordering = [
        "-created",
    ]

    def chat_link(self, obj: TgMessage):
        return SafeString(f'<a href="/hsearch/hsearch/chat/{obj.chat.id}/">{obj.chat}</a>')

    chat_link.short_description = 'chat'

    def apartment_link(self, obj: TgMessage):
        return SafeString(f'<a href="/hsearch/hsearch/apartment/{obj.apartment.id}/">{obj.apartment}</a>')

    apartment_link.short_description = "Apartment"
