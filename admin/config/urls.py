from django.conf import settings
from django.conf.urls.static import static
from django.urls import path
from django.views.generic import RedirectView

from hsearch.admin import admin

admin.autodiscover()

urlpatterns = [
    path('', RedirectView.as_view(url='/hsearch/')),
    path('hsearch/', admin.site.urls),
    *static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
]
