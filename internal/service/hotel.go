package service

import (
	"errors"

	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type HotelService struct {
	hotels  *repository.HotelRepository
	catalog *repository.HotelCatalogRepository
}

func NewHotelService(hotels *repository.HotelRepository, catalog *repository.HotelCatalogRepository) *HotelService {
	return &HotelService{hotels: hotels, catalog: catalog}
}

func (s *HotelService) GetBySlug(slug, refCode string) (*dto.HotelResponse, error) {
	hotel, err := s.hotels.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	if refCode != "" && refCode != hotel.Code && refCode != hotel.ReferralCode {
		return nil, apperrors.New(apperrors.ErrNotFound.Code, "unrecognised hotel referral code", apperrors.ErrNotFound.Status)
	}
	return toHotelResponse(hotel), nil
}

func (s *HotelService) GetByCode(code string) (*dto.HotelResponse, error) {
	hotel, err := s.hotels.FindByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	return toHotelResponse(hotel), nil
}

func (s *HotelService) ListProductsByCode(code string, featuredOnly bool) ([]dto.ProductResponse, error) {
	hotel, err := s.hotels.FindByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	var products []models.Product
	if featuredOnly {
		products, err = s.hotels.ListFeaturedProducts(hotel.ID)
	} else {
		products, err = s.hotels.ListProducts(hotel.ID)
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list products", apperrors.ErrInternal.Status)
	}
	return toProductResponses(products), nil
}

func (s *HotelService) ListProducts(slug string, featuredOnly bool) ([]dto.ProductResponse, error) {
	hotel, err := s.hotels.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	var products []models.Product
	if featuredOnly {
		products, err = s.hotels.ListFeaturedProducts(hotel.ID)
	} else {
		products, err = s.hotels.ListProducts(hotel.ID)
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list products", apperrors.ErrInternal.Status)
	}
	return toProductResponses(products), nil
}

func (s *HotelService) PartnersLanding(catalogSlug string) (*dto.PartnersLandingResponse, error) {
	hotels, err := s.hotels.ListActive()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list hotels", apperrors.ErrInternal.Status)
	}

	partnerCards := make([]dto.PartnerHotelCard, 0)
	partnerCount := 0
	for _, hotel := range hotels {
		if hotel.Slug == "necha-demo" {
			continue
		}
		partnerCount++
		partnerCards = append(partnerCards, dto.PartnerHotelCard{
			Name:     hotel.Name,
			Location: hotel.Location,
			Initials: hotel.Initials,
			Slug:     hotel.Slug,
		})
	}

	taken := partnerCount
	if taken > dto.FoundingCohortSize {
		taken = dto.FoundingCohortSize
	}

	featured := []dto.ProductResponse{}
	if catalogSlug != "" {
		if products, listErr := s.ListProducts(catalogSlug, true); listErr == nil {
			featured = products
			if len(featured) > 4 {
				featured = featured[:4]
			}
		}
	}

	return &dto.PartnersLandingResponse{
		FoundingSpotsTotal: dto.FoundingCohortSize,
		FoundingSpotsTaken: taken,
		PartnerHotels:      partnerCards,
		FeaturedProducts:   featured,
		ActiveHotelCount:   partnerCount,
	}, nil
}

func (s *HotelService) GetProduct(slug, productSlug string) (*dto.ProductResponse, error) {
	hotel, err := s.hotels.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	product, err := s.hotels.FindProductBySlug(hotel.ID, productSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "product not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load product", apperrors.ErrInternal.Status)
	}
	resp := toProductResponses([]models.Product{*product})
	return &resp[0], nil
}

func (s *HotelService) ListRooms(slug string) ([]dto.HotelRoomResponse, error) {
	hotel, err := s.hotels.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	rooms, err := s.catalog.ListRooms(hotel.ID, true)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list rooms", apperrors.ErrInternal.Status)
	}
	result := make([]dto.HotelRoomResponse, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, dto.HotelRoomResponse{
			RoomNumber: room.RoomNumber,
			RoomType:   room.RoomType,
			Floor:      room.Floor,
		})
	}
	return result, nil
}

func (s *HotelService) ListMenu(slug string) (*dto.HotelMenuResponse, error) {
	hotel, err := s.hotels.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	categories, err := s.catalog.ListCategories(hotel.ID, models.CategoryKindMenu, true)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list menu categories", apperrors.ErrInternal.Status)
	}
	items, err := s.catalog.ListMenuItems(hotel.ID, true)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list menu items", apperrors.ErrInternal.Status)
	}
	resp := &dto.HotelMenuResponse{
		Categories: make([]dto.HotelMenuCategoryResponse, 0, len(categories)+1),
		Items:      make([]dto.HotelMenuItemResponse, 0, len(items)),
	}
	resp.Categories = append(resp.Categories, dto.HotelMenuCategoryResponse{ID: "all", Label: "All"})
	for _, cat := range categories {
		resp.Categories = append(resp.Categories, dto.HotelMenuCategoryResponse{
			ID:    cat.Slug,
			Label: cat.Label,
		})
	}
	for _, item := range items {
		resp.Items = append(resp.Items, dto.HotelMenuItemResponse{
			ID:          item.Slug,
			Category:    item.Category,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Tag:         item.Tag,
		})
	}
	return resp, nil
}

func toHotelResponse(hotel *models.Hotel) *dto.HotelResponse {
	services := []string(hotel.Services)
	if services == nil {
		services = []string{}
	}
	return &dto.HotelResponse{
		ID:           hotel.ID.String(),
		Code:         hotel.Code,
		Slug:         hotel.Slug,
		Name:         hotel.Name,
		Description:  hotel.Description,
		Address:      hotel.Address,
		City:         hotel.City,
		Location:     hotel.Location,
		Country:      hotel.Country,
		Zone:         hotel.Zone,
		Phone:        hotel.Phone,
		Initials:     hotel.Initials,
		LogoURL:      hotel.LogoURL,
		ReferralCode: hotel.ReferralCode,
		Services:     services,
		IsVerified:   hotel.IsVerified,
	}
}

func toProductResponses(products []models.Product) []dto.ProductResponse {
	result := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		result = append(result, dto.ProductResponse{
			ID:          p.ID.String(),
			Slug:        p.Slug,
			BrandName:   p.BrandName,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Badge:       p.Badge,
			Price:       p.Price,
			Currency:    p.Currency,
			ImageURL:    p.ImageURL,
			Stock:       p.Stock,
			IsFeatured:  p.IsFeatured,
		})
	}
	return result
}
