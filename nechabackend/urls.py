from django.contrib import admin
from django.urls import path,include
from django.conf import settings
from django.conf.urls.static import static




urlpatterns = [
    path("admin/", admin.site.urls),
    path('api/v1/product/', include('product.urls')),
    path('api/v1/', include('category.urls')),
    path('api/v1/', include('djoser.urls')),
    path('api/v1/', include('djoser.urls.jwt')),

] + static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)+ static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
