from django.db import models
from django.conf import settings
from product.models import Product

class ProductStock(models.Model):
    product = models.OneToOneField(Product, on_delete=models.CASCADE)
    quantity = models.PositiveIntegerField(default=0)
    
    def check_low_stock():
        products = ProductStock.objects.filter(quantity__lt=10)  # Adjust the threshold as needed
        for product_stock in products:
            pass
        # Send a notification, e.g., through email or a notification system.
        # You can use third-party packages like Django's Email Sending Framework or a notification library like django-notifications.
        # You may want to schedule this function to run periodically using a task scheduler like Celery.


class RestockingEvent(models.Model):
    product_stock = models.ForeignKey(ProductStock, on_delete=models.CASCADE)
    restocked_quantity = models.PositiveIntegerField()
    restocked_date = models.DateTimeField(auto_now_add=True)
    restocked_by = models.ForeignKey(settings.AUTH_USER_MODEL, on_delete=models.SET_NULL, null=True)
