from django.urls import path

from hsearch.views.v1 import advertisement_list

urlpatterns = [
    path("advertisement/list/", advertisement_list),
]
