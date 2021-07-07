from django.urls import path

from hsearch.views.v1 import apartment_list

urlpatterns = [
    path("apartment/list/", apartment_list),
]
