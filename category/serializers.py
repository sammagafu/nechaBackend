from rest_framework import serializers
from .models import Category,Subcategory


class SubCategorySerializer(serializers.ModelSerializer):
    class Meta:
        model = Subcategory
        fields = '__all__'


class CategorySerializer(serializers.ModelSerializer):
    subcategory = SubCategorySerializer(many=True, read_only=True)
    class Meta:
        model = Category
        fields = '__all__'