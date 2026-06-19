package service

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type ImportService struct {
	hotels   *repository.HotelRepository
	catalog  *repository.HotelCatalogRepository
}

func NewImportService(hotels *repository.HotelRepository, catalog *repository.HotelCatalogRepository) *ImportService {
	return &ImportService{hotels: hotels, catalog: catalog}
}

func (s *ImportService) ImportRooms(hotelID uuid.UUID, r io.Reader) (*dto.ImportResult, error) {
	if _, err := s.hotels.FindByID(hotelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	result := &dto.ImportResult{Kind: "rooms"}
	rows, err := readCSVRows(r)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, err.Error(), apperrors.ErrBadRequest.Status)
	}

	for i, row := range rows {
		line := i + 2
		roomNumber := strings.TrimSpace(row["room_number"])
		if roomNumber == "" {
			result.Skipped++
			continue
		}
		room := &models.HotelRoom{
			HotelID:    hotelID,
			RoomNumber: roomNumber,
			RoomType:   strings.TrimSpace(row["room_type"]),
			Floor:      strings.TrimSpace(row["floor"]),
			Notes:      strings.TrimSpace(row["notes"]),
			IsActive:   parseBoolDefault(row["is_active"], true),
		}
		created, err := s.catalog.UpsertRoom(room)
		if err != nil {
			result.Errors = append(result.Errors, formatImportError(line, err))
			continue
		}
		if created {
			result.Created++
		} else {
			result.Updated++
		}
	}
	return result, nil
}

func (s *ImportService) ImportCategories(hotelID uuid.UUID, r io.Reader) (*dto.ImportResult, error) {
	if _, err := s.hotels.FindByID(hotelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	result := &dto.ImportResult{Kind: "categories"}
	rows, err := readCSVRows(r)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, err.Error(), apperrors.ErrBadRequest.Status)
	}

	for i, row := range rows {
		line := i + 2
		slug := slugify(row["slug"])
		label := strings.TrimSpace(row["label"])
		if slug == "" || label == "" {
			result.Skipped++
			continue
		}
		kind := models.CategoryKind(strings.ToLower(strings.TrimSpace(row["kind"])))
		if kind != models.CategoryKindProduct && kind != models.CategoryKindMenu {
			result.Errors = append(result.Errors, formatImportError(line, errors.New("kind must be product or menu")))
			continue
		}
		sortOrder, _ := strconv.Atoi(strings.TrimSpace(row["sort_order"]))
		category := &models.HotelCategory{
			HotelID:   hotelID,
			Slug:      slug,
			Label:     label,
			Kind:      kind,
			SortOrder: sortOrder,
			IsActive:  parseBoolDefault(row["is_active"], true),
		}
		created, err := s.catalog.UpsertCategory(category)
		if err != nil {
			result.Errors = append(result.Errors, formatImportError(line, err))
			continue
		}
		if created {
			result.Created++
		} else {
			result.Updated++
		}
	}
	return result, nil
}

func (s *ImportService) ImportMenu(hotelID uuid.UUID, r io.Reader) (*dto.ImportResult, error) {
	if _, err := s.hotels.FindByID(hotelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	result := &dto.ImportResult{Kind: "menu"}
	rows, err := readCSVRows(r)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, err.Error(), apperrors.ErrBadRequest.Status)
	}

	for i, row := range rows {
		line := i + 2
		name := strings.TrimSpace(row["name"])
		category := slugify(row["category"])
		if name == "" || category == "" {
			result.Skipped++
			continue
		}
		slug := slugify(row["slug"])
		if slug == "" {
			slug = slugify(name)
		}
		price, err := parsePrice(row["price"])
		if err != nil {
			result.Errors = append(result.Errors, formatImportError(line, err))
			continue
		}
		sortOrder, _ := strconv.Atoi(strings.TrimSpace(row["sort_order"]))
		currency := strings.TrimSpace(row["currency"])
		if currency == "" {
			currency = "TZS"
		}
		item := &models.HotelMenuItem{
			HotelID:     hotelID,
			Slug:        slug,
			Category:    category,
			Name:        name,
			Description: strings.TrimSpace(row["description"]),
			Price:       price,
			Currency:    currency,
			Tag:         strings.TrimSpace(row["tag"]),
			SortOrder:   sortOrder,
			IsActive:    parseBoolDefault(row["is_active"], true),
		}
		created, err := s.catalog.UpsertMenuItem(item)
		if err != nil {
			result.Errors = append(result.Errors, formatImportError(line, err))
			continue
		}
		if created {
			result.Created++
		} else {
			result.Updated++
		}
	}
	return result, nil
}

func readCSVRows(r io.Reader) ([]map[string]string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, errors.New("csv must include a header row and at least one data row")
	}
	headers := normalizeHeaders(records[0])
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		if isEmptyRecord(record) {
			continue
		}
		row := make(map[string]string, len(headers))
		for i, header := range headers {
			if header == "" {
				continue
			}
			if i < len(record) {
				row[header] = strings.TrimSpace(record[i])
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, errors.New("csv has no data rows")
	}
	return rows, nil
}

func normalizeHeaders(headers []string) []string {
	out := make([]string, len(headers))
	for i, h := range headers {
		out[i] = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(h, " ", "_")))
	}
	return out
}

func isEmptyRecord(record []string) bool {
	for _, cell := range record {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func parseBoolDefault(raw string, fallback bool) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}

func parsePrice(raw string) (int64, error) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, ",", ""))
	if raw == "" {
		return 0, errors.New("price is required")
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.New("invalid price")
	}
	if value < 0 {
		return 0, errors.New("price must be zero or positive")
	}
	return value, nil
}

func slugify(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, " ", "-")
	raw = strings.ReplaceAll(raw, "_", "-")
	return raw
}

func formatImportError(line int, err error) string {
	return "row " + strconv.Itoa(line) + ": " + err.Error()
}
