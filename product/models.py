from django.db import models
from django.utils.translation import ugettext_lazy as _
from django.utils.text import slugify
from django.urls import reverse
from django.conf import settings
from django.dispatch import receiver
from django.db.models.signals import post_save
from django.utils import timezone
import uuid 

class Product(models.Model):
    name = models.CharField(max_length=255)
    description = models.TextField()
    price = models.DecimalField(max_digits=10, decimal_places=2)
    category = models.ForeignKey('category.Category', on_delete=models.CASCADE)
    cover = models.ImageField(upload_to='products/', null=True, blank=True)

    def __str__(self):
        return self.name
    
class ProductImage(models.Model):
    product = models.ForeignKey(Product, on_delete=models.CASCADE)
    image = models.ImageField(upload_to='product/images/')

    
# class Product(models.Model):
#     product_quantity = [
#         ('KG','Kilogram'),
#         ('L','Litres'),
#         ('DZ','Dozens')
#     ]

#     name = models.CharField(verbose_name=_('Enter product name'), max_length=100)
#     selling_price = models.FloatField(verbose_name=_('Selling Price'),null=False,help_text="Unit Selling Price")
#     slug = models.SlugField(null=True,blank=False,unique=True,editable=False)
#     cover = models.ImageField(verbose_name=_('Product Picture'),upload_to='product/covers/',help_text=_("Must be 800px by 800px"))
#     si_unit = models.CharField(verbose_name=_('Product SI Unit'),choices=product_quantity, max_length=3,default="KG")
#     quantity = models.IntegerField(verbose_name=_("Number of Products In"))
#     sku = models.CharField(verbose_name=_("Stock Keeping Unit"),editable=False,max_length=12)
#     description = models.TextField()
#     discount = models.FloatField(verbose_name=_('Discount'),null=True,blank=True,help_text="Discount Percentage, without a percentage sign")
#     startsAt = models.DateTimeField(verbose_name=_("Sales Start Date"), auto_now=False, auto_now_add=False,blank=True,null=True)
#     endsAt = models.DateTimeField(verbose_name=_("Sales Ends Date"), auto_now=False, auto_now_add=False,blank=True,null=True)
#     created_at = models.DateTimeField(verbose_name=_("Product Created Date"), auto_now_add=True)
#     updated_at = models.DateTimeField(verbose_name=_("Product Created Date"), auto_now=True)
#     low_amount  = models.IntegerField(verbose_name=_("Low amount"),null=True)
#     # stocking starts here
    


#     # TODO: Define fields here

#     class Meta:
#         verbose_name = 'Product'
#         verbose_name_plural = 'Products'

#     def __str__(self):
#         return self.name

#     def save(self):
#         self.sku = "phema-" + uuid.uuid4().hex[:6].lower()
#         self.slug = slugify(self.name)
#         super(Product,self).save()

#     def get_absolute_url(self):
#         """Return absolute url for Product."""
#         return reverse('product:detail', args=[str(self.slug)])

class ProductReview(models.Model):     
    product = models.ForeignKey(Product, verbose_name=_("Review"), on_delete=models.CASCADE)
    user = models.ForeignKey(settings.AUTH_USER_MODEL, verbose_name="Reviwing User", on_delete=models.CASCADE)
    review = models.TextField()
    class Meta:
        verbose_name = 'Product Review'
        verbose_name_plural = 'Product Reviews'

    def __str__(self):
        """Unicode representation of ProductReview."""
        pass