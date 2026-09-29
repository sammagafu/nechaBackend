package models

import "testing"

func TestNormalizeInquiryType(t *testing.T) {
	cases := map[string]string{
		InquiryTypeHotelPartner:          InquiryTypeHotelPartnership,
		InquiryTypeBrandPartner:          InquiryTypeBrandOnboarding,
		InquiryTypeAffiliatePartner:      InquiryTypePartnerReferralSignup,
		InquiryTypePartnerReferral:       InquiryTypePartnerReferralSignup,
		InquiryTypeHotelPartnership:      InquiryTypeHotelPartnership,
		InquiryTypeBrandOnboarding:       InquiryTypeBrandOnboarding,
		InquiryTypeEventListing:          InquiryTypeEventListing,
		InquiryTypePartnerReferralSignup: InquiryTypePartnerReferralSignup,
		InquiryTypeContact:               InquiryTypeContact,
	}
	for in, want := range cases {
		if got := NormalizeInquiryType(in); got != want {
			t.Fatalf("NormalizeInquiryType(%q) = %q, want %q", in, got, want)
		}
	}
}
