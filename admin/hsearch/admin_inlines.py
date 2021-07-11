from django.contrib import admin
from django.db import models
from django.forms import Widget
from django.utils.safestring import SafeString

from hsearch.models import Answer, Feedback, Image


class BaseReadOnly(admin.TabularInline):
    extra = 0
    classes = [
        "collapse",
    ]

    def has_delete_permission(self, request, obj=None):
        return False

    def has_add_permission(self, request, obj):
        return False

    def has_change_permission(self, request, obj=None):
        return False


class AdminImageWidget(Widget):
    def render(self, name, value, attrs=None, renderer=None):
        display_name = value.split("/")[-1]
        img = f'<img height="200px" src="{value}" alt="{display_name}"/>'
        return SafeString(f'<a href="{value}" target="_blank">{img}</a>')


class FeedbackInline(BaseReadOnly):
    model = Feedback
    fields = [
        "body",
        "created",
    ]


class AnswerInline(BaseReadOnly):
    model = Answer
    fields = [
        "apartment",
        "dislike",
        "created",
    ]


class ImageInline(BaseReadOnly):
    model = Image
    fields = [
        "path",
        "created",
    ]
    readonly_fields = [
        "created",
    ]
    formfield_overrides = {
        models.CharField: {"widget": AdminImageWidget}
    }

    def has_change_permission(self, request, obj=None):
        return True
