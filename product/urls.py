from django.urls import path
from . import views


urlpatterns = [
    path('',views.ProductList.as_view(),name="product-api-list"),
    path('<slug>/',views.ProductDetail.as_view(),name="product-api-detail")
]