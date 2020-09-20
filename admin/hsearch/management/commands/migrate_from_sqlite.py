from django.core.management import BaseCommand

from hsearch.models import Chat, Offer, Answer, Feedback, Image, TgMessage


class Command(BaseCommand):
    help = 'Migrate hsearch project from sqlite to postgres'

    def handle(self, *args, **options):
        Chat.objects.using('default').all().delete()
        Chat.objects.using('default').bulk_create(Chat.objects.using('sqlite').all())
        print('Done Chat')

        Offer.objects.using('default').all().delete()
        Offer.objects.using('default').bulk_create(Offer.objects.using('sqlite').all())
        print('Done Offer')

        Answer.objects.using('default').all().delete()
        Answer.objects.using('default').bulk_create(Answer.objects.using('sqlite').all())
        print('Done Answer')

        Feedback.objects.using('default').all().delete()
        Feedback.objects.using('default').bulk_create(Feedback.objects.using('sqlite').all())
        print('Done Feedback')

        Image.objects.using('default').all().delete()
        Image.objects.using('default').bulk_create(Image.objects.using('sqlite').all())
        print('Done Image')

        TgMessage.objects.using('default').all().delete()
        TgMessage.objects.using('default').bulk_create(TgMessage.objects.using('sqlite').all())
        print('Done TgMessage')
