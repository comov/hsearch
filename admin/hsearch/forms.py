from captcha.fields import ReCaptchaField
from django.contrib.admin.forms import AdminAuthenticationForm as _AdminAuthenticationForm


class AdminAuthenticationForm(_AdminAuthenticationForm):
    captcha = ReCaptchaField()
