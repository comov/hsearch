from django.contrib.auth.models import User
from django.db import models
from django.db.models.signals import post_save
from django.dispatch import receiver
from unixtimestampfield.fields import UnixTimeStampField


class Chat(models.Model):
    PRIVATE = "private"
    SUPERGROUP = "supergroup"
    TYPE_CHOICES = (
        (PRIVATE, "private"),
        (SUPERGROUP, "supergroup"),
    )
    user = models.OneToOneField(to="auth.User", on_delete=models.CASCADE, null=True, related_name="chat")
    chat_id = models.BigIntegerField(default=0, blank=True)
    username = models.CharField(max_length=100, default="", blank=True)
    title = models.CharField(max_length=100, default="", blank=True)
    c_type = models.CharField(max_length=20, choices=TYPE_CHOICES, default=PRIVATE)
    created = UnixTimeStampField()
    enable = models.BooleanField(default=True)
    diesel = models.BooleanField(default=True)
    lalafo = models.BooleanField(default=True)
    house = models.BooleanField(default=True)
    photo = models.BooleanField(default=True)
    usd = models.CharField(max_length=100, default="0:0")
    kgs = models.CharField(max_length=100, default="0:0")

    def __str__(self):
        name = self.title if self.title else self.username
        return f"{name} {self.get_c_type_display()!r} (#{self.id})"


class Image(models.Model):
    apartment = models.ForeignKey("hsearch.Apartment", on_delete=models.CASCADE, related_name="images")
    path = models.CharField(max_length=255, default="", unique=True)
    created = UnixTimeStampField()

    def __str__(self):
        return f"{self.apartment.topic} ({self.path})"

    available_fields = [
        "apartment_id",
        "path",
        "created",
    ]

    def to_dict(self, fields_list: list) -> dict:
        return {field: getattr(self, field, None) for field in fields_list}


class Apartment(models.Model):
    DIESEL = "diesel"
    LALAFO = "lalafo"
    HOUSE = "house"
    SITE_CHOICES = (
        (DIESEL, "diesel"),
        (LALAFO, "lalafo"),
        (HOUSE, "house"),
    )

    CURRENCY_CHOICE = (
        (0, "-"),
        (1, "USD"),
        (2, "KGS"),
    )

    external_id = models.BigIntegerField()
    url = models.CharField(max_length=255, default="")
    topic = models.CharField(max_length=255, default="")
    phone = models.CharField(max_length=255, default="", blank=True)
    rooms = models.IntegerField(default=0, blank=True)
    body = models.TextField(default="", blank=True)
    images_count = models.IntegerField(default=0)
    price = models.IntegerField(default=0, blank=True)
    currency = models.IntegerField(default=0, blank=True)
    area = models.IntegerField(default=0, blank=True)
    city = models.CharField(max_length=100, default="", blank=True)
    room_type = models.CharField(max_length=100, default="", blank=True)
    site = models.CharField(max_length=20, default="", choices=SITE_CHOICES)
    floor = models.IntegerField(default=0, blank=True)
    max_floor = models.IntegerField(default=0, blank=True)
    district = models.CharField(max_length=100, default="", blank=True)
    lat = models.FloatField(default=0.0, blank=True)
    lon = models.FloatField(default=0.0, blank=True)
    created = UnixTimeStampField()

    def __str__(self):
        return f"{self.topic} {self.get_site_display()!r} (#{self.id})"

    def save(self, force_insert=False, force_update=False, using=None, update_fields=None):
        self.images_count = self.images.count()
        super().save(force_insert, force_update, using, update_fields)

    available_fields = [
        "id",
        "url",
        "topic",
        "phone",
        "room_numbers",
        "body",
        "images_count",
        "price",
        "currency",
        "area",
        "city",
        "room_type",
        "site",
        "floor",
        "district",
        "created",
    ]

    available_relations_fields = [f"images__{i}" for i in Image.available_fields]

    def to_dict(self, fields_list: list, relations: list, relations_fields: list) -> dict:
        _dict_object = {field: getattr(self, field, None) for field in fields_list}
        for relation in relations:
            _dict_object.setdefault(relation, [])
            for relation_item in getattr(self, relation).only(*relations_fields).all():
                _dict_object[relation].append(relation_item.to_dict(relations_fields))
        return _dict_object


class Answer(models.Model):
    chat = models.ForeignKey("hsearch.Chat", on_delete=models.CASCADE, related_name="answers")
    apartment = models.ForeignKey("hsearch.Apartment", on_delete=models.CASCADE, related_name="answers")
    dislike = models.BooleanField(default=False)
    created = UnixTimeStampField()

    def __str__(self):
        return f"{self.chat_id} => {self.apartment_id} ({self.dislike})"


class Feedback(models.Model):
    username = models.CharField(max_length=100, default="")
    chat = models.ForeignKey("hsearch.Chat", on_delete=models.CASCADE, related_name="feedbacks")
    body = models.TextField(default="")
    created = UnixTimeStampField()

    def __str__(self):
        return self.username if self.username else self.chat


class TgMessage(models.Model):
    APARTMENT = "apartment"
    PHOTO = "photo"
    DESCRIPTION = "description"
    KIND_CHOICES = (
        (APARTMENT, "Apartment"),
        (PHOTO, "Photo"),
        (DESCRIPTION, "Description"),
    )
    created = UnixTimeStampField()
    message_id = models.IntegerField(default=0)
    apartment = models.ForeignKey("hsearch.Apartment", on_delete=models.CASCADE, related_name="messages")
    chat = models.ForeignKey("hsearch.Chat", on_delete=models.CASCADE, related_name="messages")
    kind = models.CharField(max_length=50, choices=KIND_CHOICES, default=APARTMENT)


@receiver(post_save, sender=User)
def create_chat(sender, instance: User, created: bool, **kwargs):
    if not created:
        return
    Chat.objects.create(user=instance)
