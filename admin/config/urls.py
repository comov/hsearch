from django.conf import settings
from django.conf.urls.static import static
from django.urls import include, path
from django.views.decorators.csrf import csrf_exempt
from graphene_django.views import GraphQLView

from hsearch.admin import admin
from hsearch.views import index_page

admin.autodiscover()

urlpatterns = [
    path("", index_page),
    path("auth/", include("social_django.urls", namespace="social")),
    path("hsearch/", admin.site.urls),
    path("graphql/", csrf_exempt(GraphQLView.as_view(graphiql=True))),
    *static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
]
