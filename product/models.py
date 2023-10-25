from django.db import models
from django.utils.translation import ugettext_lazy as _
from django.utils.text import slugify
from django.conf import settings
from django_resized import ResizedImageField



import uuid 

class Product(models.Model):
    name = models.CharField(max_length=255)
    slug = models.SlugField(null=True,blank=False,unique=True,editable=False)
    description = models.TextField()
    price = models.DecimalField(max_digits=10, decimal_places=2)
    category = models.ForeignKey('category.Category', on_delete=models.CASCADE)
    cover = ResizedImageField(upload_to = 'product/images/%Y/%m/%d',verbose_name=_("Products image"),size=[900, 1000], crop=['middle', 'center'],null=True,blank=True)

    # cover = models.ImageField(upload_to='products/', null=True, blank=True)

    def __str__(self):
        return self.name
    
    def save(self, *args, **kwargs):
        if not self.slug:
            self.slug = slugify(self.name)
        self.slug = slugify(self.name)
        return super().save()

    
class ProductImage(models.Model):
    product = models.ForeignKey(Product, on_delete=models.CASCADE)
    image = ResizedImageField(upload_to = 'product/images/%Y/%m/%d',verbose_name=_("Products image"),size=[900, 1000], crop=['middle', 'center'],null=True,blank=True)

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