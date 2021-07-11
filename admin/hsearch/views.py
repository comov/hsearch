from django.conf import settings
from django.shortcuts import render


def index_page(request):
    return render(request, "index.html", context={
        "bot_name": settings.TG_NAME,
        "auth_url": f"{request.scheme}://{request.get_host()}{settings.TG_LOGIN_REDIRECT_URL}",
    })
