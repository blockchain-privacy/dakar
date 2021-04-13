package constants

// MixingTypes is the set of all mixing denomination types
var MixingTypes = [5]int{PrivacyMixing0, PrivacyMixing1, PrivacyMixing2, PrivacyMixing3, PrivacyMixing4}

// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// privacy type ranges
const (
	PrivacyMixing0    = 5  // 10.0001 -- 1000010000
	PrivacyMixing1    = 10 // 01.00001 -- 100001000
	PrivacyMixing2    = 15 // 00.100001 -- 10000100
	PrivacyMixing3    = 20 // 00.0100001 -- 1000010
	PrivacyMixing4    = 25 // 00.00100001 -- 100001
	PrivacyMixingLast = 99 // the maximum id in the privacy mixing range (0 - 99)

	PrivacyDestinationFirst = 100
	PrivacyDestination      = 101
	PrivacyDestinationLast  = 199

	PrivacyOriginFirst = 200
	PrivacyOrigin      = 201
	PrivacyOriginLast  = 299

	PrivacyCollateralCreationFirst = 300
	PrivacyCollateralCreation      = 301
	PrivacyCollateralCreationLast  = 399

	PrivacyCollateralPaymentFirst = 400
	PrivacyCollateralPayment      = 401
	PrivacyCollateralPaymentLast  = 499
)

// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// string representation of privacy type ranges for performance
const (
	StrPrivacyMixing0    = "5"
	StrPrivacyMixing1    = "10"
	StrPrivacyMixing2    = "15"
	StrPrivacyMixing3    = "20"
	StrPrivacyMixing4    = "25"
	StrPrivacyMixingLast = "99"

	StrPrivacyDestinationFirst = "100"
	StrPrivacyDestination      = "101"
	StrPrivacyDestinationLast  = "199"

	StrPrivacyOriginFirst = "200"
	StrPrivacyOrigin      = "201"
	StrPrivacyOriginLast  = "299"

	StrPrivacyCollateralCreationFirst = "300"
	StrPrivacyCollateralCreation      = "301"
	StrPrivacyCollateralCreationLast  = "399"

	StrPrivacyCollateralPaymentFirst = "400"
	StrPrivacyCollateralPayment      = "401"
	StrPrivacyCollateralPaymentLast  = "499"
)
