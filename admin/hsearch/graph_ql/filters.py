import django_filters
from graphene import DateTime
from graphene_django.converter import convert_django_field
from unixtimestampfield import UnixTimeStampField

from hsearch.models import Apartment, Image


@convert_django_field.register(UnixTimeStampField)
def convert_field_to_string(field, registry=None):
    return DateTime(description=field.help_text, required=not field.null)


class ApartmentFilter(django_filters.FilterSet):
    created = django_filters.DateTimeFilter()

    topic = django_filters.CharFilter(lookup_expr="icontains")
    body = django_filters.CharFilter(lookup_expr="icontains")

    rooms = django_filters.Filter(method="filter_range")
    area = django_filters.Filter(method="filter_range")
    floor = django_filters.Filter(method="filter_range")
    price = django_filters.Filter(method="filter_range")

    with_images = django_filters.BooleanFilter(method="filter_with_images")

    @staticmethod
    def filter_range(queryset, name, value: str):
        if len(value.split(",")) == 2:
            return queryset.filter(**{f"{name}__range": value.split(",")})
        elif value.isdigit():
            return queryset.filter(**{name: value})
        return queryset.none()

    @staticmethod
    def filter_with_images(queryset, name, value: bool):
        if value:
            return queryset.filter(images_count__gt=0)
        return queryset.filter(images_count__lt=1)

    class Meta:
        model = Apartment
        fields = [
            "id",
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


class ImageFilter(django_filters.FilterSet):
    created = django_filters.DateTimeFilter()

    class Meta:
        model = Image
        fields = [
            "id",
            "apartment",
            "path",
            "created",
        ]
