from django.conf import settings
from django.conf.urls.static import static
from django.urls import include, path
from django.views.generic import RedirectView
from graphene_django.views import GraphQLView

from hsearch.admin import admin

admin.autodiscover()

urlpatterns = [
    path('', RedirectView.as_view(url='/hsearch/')),
    path('hsearch/', admin.site.urls),
    path("v1/", include("hsearch.urls_v1")),
    path("graphql/", GraphQLView.as_view(graphiql=True)),
    *static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
]
