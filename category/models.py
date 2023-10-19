from django.db import models
from django.utils.text import slugify

# Create your models here.
class Category(models.Model):
    name = models.CharField(max_length=100)
    slug = models.SlugField(null=True,blank=False,unique=True,editable=False)

    def __str__(self):
        return self.name
    
    def save(self):
        self.slug = slugify(self.name)
        super(Category,self).save()

class Subcategory(models.Model):
    name = models.CharField(max_length=100)
    slug = models.SlugField(null=True,blank=False,unique=True,editable=False)
    category = models.ForeignKey(Category, on_delete=models.CASCADE, related_name='subcategories')

    def __str__(self):
        return self.name
    
    def save(self):
        self.slug = slugify(self.name)
        super(Category,self).save()
    




    

    