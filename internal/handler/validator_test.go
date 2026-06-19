package handler

import "testing"

type sliceFixture struct {
	Items []string `json:"items" validate:"required,min=1,dive,required"`
}

type minIntFixture struct {
	Count int `json:"count" validate:"required,min=2"`
}

func TestValidateStructSliceMinAndDive(t *testing.T) {
	if err := validateStruct(&sliceFixture{Items: []string{"a"}}); err != nil {
		t.Fatalf("expected valid slice, got %v", err)
	}
	if err := validateStruct(&sliceFixture{Items: []string{}}); err == nil {
		t.Fatal("expected min=1 validation error for empty slice")
	}
}

func TestValidateStructMinInt(t *testing.T) {
	if err := validateStruct(&minIntFixture{Count: 2}); err != nil {
		t.Fatalf("expected valid count, got %v", err)
	}
	if err := validateStruct(&minIntFixture{Count: 1}); err == nil {
		t.Fatal("expected min=2 validation error")
	}
}
