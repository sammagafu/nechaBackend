from rest_framework.routers import DefaultRouter
from .views import CategoryViewSet, SubcategoryViewSet
from django.urls import path, include


router = DefaultRouter()
router.register(r'categories', CategoryViewSet)
router.register(r'subcategories', SubcategoryViewSet)


urlpatterns = [
    path('', include(router.urls)),
]






