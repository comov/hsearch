from graphene import relay
from graphene_django import DjangoObjectType

from hsearch.graph_ql.filters import ApartmentFilter, ImageFilter
from hsearch.models import Apartment, Image


class ApartmentNode(DjangoObjectType):
    class Meta:
        model = Apartment
        filterset_class = ApartmentFilter
        interfaces = (relay.Node,)


class ImageNode(DjangoObjectType):
    class Meta:
        model = Image
        filterset_class = ImageFilter
        interfaces = (relay.Node,)
