from graphene import ObjectType, Schema, relay
from graphene_django.filter import DjangoFilterConnectionField

from hsearch.graph_ql.schema import ApartmentNode, ImageNode


class Query(ObjectType):
    apartment = relay.Node.Field(ApartmentNode)
    all_apartments = DjangoFilterConnectionField(ApartmentNode)

    image = relay.Node.Field(ImageNode)
    all_images = DjangoFilterConnectionField(ImageNode)


schema = Schema(query=Query)
