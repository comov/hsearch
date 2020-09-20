from django.db import models


class Chat(models.Model):
    PRIVATE = 'private'
    SUPERGROUP = 'supergroup'
    TYPE_CHOICES = (
        (PRIVATE, 'private'),
        (SUPERGROUP, 'supergroup'),
    )
    id = models.BigIntegerField(default=0, unique=True, null=False, primary_key=True)
    username = models.CharField(max_length=100, default='')
    title = models.CharField(max_length=100, default='')
    c_type = models.CharField(max_length=20, choices=TYPE_CHOICES, default=PRIVATE)
    created = models.IntegerField(default=0)
    enable = models.BooleanField(default=True)
    diesel = models.BooleanField(default=True)
    lalafo = models.BooleanField(default=True)
    house = models.BooleanField(default=True)
    photo = models.BooleanField(default=True)
    usd = models.CharField(max_length=100, default='0:0')
    kgs = models.CharField(max_length=100, default='0:0')

    class Meta:
        db_table = 'chat'
        managed = False


class Offer(models.Model):
    id = models.IntegerField(default=0, unique=True, null=False, primary_key=True)
    created = models.IntegerField(default=0)
    url = models.CharField(max_length=255, default='')
    topic = models.CharField(max_length=255, default='')
    full_price = models.CharField(max_length=50, default='')
    phone = models.CharField(max_length=255, default='')
    room_numbers = models.CharField(max_length=255, default='')
    body = models.TextField(default='')
    images = models.IntegerField(default=0)
    price = models.IntegerField(default=0)
    currency = models.CharField(max_length=10, default='')
    area = models.CharField(max_length=100, default='')
    city = models.CharField(max_length=100, default='')
    room_type = models.CharField(max_length=100, default='')
    site = models.CharField(max_length=20, default='')
    floor = models.CharField(max_length=20, default='')
    district = models.CharField(max_length=100, default='')

    class Meta:
        db_table = 'offer'
        managed = False


class Answer(models.Model):
    created = models.IntegerField(default=0)
    chat = models.BigIntegerField(default=0, db_index=True)
    offer_id = models.IntegerField(default=0, db_index=True)
    dislike = models.BooleanField(default=False)

    class Meta:
        db_table = 'answer'
        managed = False


class Feedback(models.Model):
    created = models.IntegerField(default=0)
    chat = models.BigIntegerField(default=0, db_index=True)
    username = models.CharField(max_length=100, default='')
    body = models.TextField(default='')

    class Meta:
        db_table = 'feedback'
        managed = False


class Image(models.Model):
    offer_id = models.IntegerField(default=0, db_index=True)
    path = models.CharField(max_length=255, default='', unique=True)
    created = models.IntegerField(default=0)

    class Meta:
        db_table = 'image'
        managed = False


class TgMessage(models.Model):
    OFFER = 'offer'
    PHOTO = 'offer'
    DESCRIPTION = 'description'
    KIND_CHOICES = (
        (OFFER, 'offer'),
        (PHOTO, 'offer'),
        (DESCRIPTION, 'description'),
    )
    created = models.IntegerField(default=0)
    message_id = models.IntegerField(default=0)
    offer_id = models.IntegerField(default=0, db_index=True)
    chat = models.BigIntegerField(default=0, db_index=True)
    kind = models.CharField(max_length=50, choices=KIND_CHOICES, default=OFFER)

    class Meta:
        db_table = 'tg_messages'
        managed = False
