from django.db import models
from django.utils.text import slugify

# Create your models here.
class Category(models.Model):
    name = models.CharField(max_length=100)
    slug = models.SlugField(null=True,blank=False,unique=True,editable=False)

    class Meta:
        verbose_name = 'Product Category'
        verbose_name_plural = 'Product Categories'

    def __str__(self):
        return self.name
    
    def save(self, *args, **kwargs):
        if not self.slug:
            self.slug = slugify(self.name)
        self.slug = slugify(self.name)
        return super().save()

class Subcategory(models.Model):
    name = models.CharField(max_length=100)
    slug = models.SlugField(null=True,blank=False,unique=True,editable=False)
    category = models.ForeignKey(Category, on_delete=models.CASCADE, related_name='subcategories')

    class Meta:
        verbose_name = 'Product Sub-category'
        verbose_name_plural = 'Product Sub-categories'

    def __str__(self):
        return self.name
    
    def save(self, *args, **kwargs):
        if not self.slug:
            self.slug = slugify(self.name)
        self.slug = slugify(self.name)
        return super().save()
    




    

    