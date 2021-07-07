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
    topic_icontains = django_filters.CharFilter(method="filter_topic_icontains")
    phone = django_filters.CharFilter(lookup_expr="istartswith")
    body = django_filters.CharFilter(lookup_expr="icontains")

    def filter_topic_icontains(self, queryset, name, value):
        return queryset.filter(topic__icontains=value)

    class Meta:
        model = Apartment
        fields = [
            "id",
            "url",
            "topic",
            "full_price",
            "phone",
            "room_numbers",
            "body",
            "images_count",
            "price",
            "currency",
            "area",
            "city",
            "room_type",
            "site",
            "floor",
            "district",
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
