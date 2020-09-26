from pathlib import Path

import environ

# BASE
# ----------------------------------------------------------------------------
BASE_DIR = Path(__file__).resolve().parent.parent

# ENVIRONMENT
# ----------------------------------------------------------------------------
env = environ.Env(
    DJANGO_DEBUG=(bool, False),
)
environ.Env.read_env(str(BASE_DIR.joinpath('../.env')))

DEBUG = env.bool('DJANGO_DEBUG', default=False)

# SECURITY
# ----------------------------------------------------------------------------
SECRET_KEY = env('DJANGO_SECRET_KEY', default='123') if DEBUG else env('DJANGO_SECRET_KEY')
ALLOWED_HOSTS = ['*'] if DEBUG else env.list('DJANGO_ALLOWED_HOSTS')

# APPLICATIONS
# ----------------------------------------------------------------------------
INSTALLED_APPS = [
    'django.contrib.admin',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.messages',
    'django.contrib.staticfiles',

    'captcha',
    'hsearch',
]

# MIDDLEWARE
# ----------------------------------------------------------------------------
MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.common.CommonMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'django.contrib.messages.middleware.MessageMiddleware',
    'django.middleware.clickjacking.XFrameOptionsMiddleware',
]

# URLS
# ----------------------------------------------------------------------------
ROOT_URLCONF = 'config.urls'
LOGOUT_REDIRECT_URL = '/'

# TEMPLATES
# ----------------------------------------------------------------------------
TEMPLATES = [
    {
        'BACKEND': 'django.template.backends.django.DjangoTemplates',
        'DIRS': [BASE_DIR / 'templates'],
        'APP_DIRS': True,
        'OPTIONS': {
            'context_processors': [
                'django.template.context_processors.debug',
                'django.template.context_processors.request',
                'django.contrib.auth.context_processors.auth',
                'django.contrib.messages.context_processors.messages',
            ],
        },
    },
]

# WSGI
# ----------------------------------------------------------------------------
WSGI_APPLICATION = 'config.wsgi.application'

# DATABASES
# ----------------------------------------------------------------------------
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': 'hsearch',
        'USER': 'hsearch',
        'PASSWORD': env('DJANGO_DB_PASSWORD', default='hsearch'),
        'HOST': env('DJANGO_DB_HOST', default='localhost'),
        'PORT': env.int('DJANGO_DB_PORT', default=5432),
    },
}

# AUTHENTICATION
# ----------------------------------------------------------------------------
AUTH_PASSWORD_VALIDATORS = [
    {
        'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator',
    },
]

# LOCALIZATION
# ----------------------------------------------------------------------------
LANGUAGE_CODE = 'en-us'
TIME_ZONE = 'UTC'
USE_I18N = True
USE_L10N = True
USE_TZ = True

# STATIC
# ----------------------------------------------------------------------------
STATIC_URL = '/static/'
STATIC_ROOT = BASE_DIR / 'static'

# django-captcha-admin
# ----------------------------------------------------------------------------
RECAPTCHA_PUBLIC_KEY = env('RECAPTCHA_PUBLIC_KEY', default='')
RECAPTCHA_PRIVATE_KEY = env('RECAPTCHA_PRIVATE_KEY', default='')
