package constants

// MixingTypes is the set of all mixing denomination types
var MixingTypes = [5]PrivacyType{PrivacyMixing0, PrivacyMixing1, PrivacyMixing2, PrivacyMixing3, PrivacyMixing4}

type PrivacyType uint16

// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// DO NOT CHANGE THE ORDER/NUMBERING OF THESE CONSTANTS
// privacy type ranges
const (
	PrivacyMixing0    = PrivacyType(5)  // 10.0001 -- 1000010000
	PrivacyMixing1    = PrivacyType(10) // 01.00001 -- 100001000
	PrivacyMixing2    = PrivacyType(15) // 00.100001 -- 10000100
	PrivacyMixing3    = PrivacyType(20) // 00.0100001 -- 1000010
	PrivacyMixing4    = PrivacyType(25) // 00.00100001 -- 100001
	PrivacyMixingLast = PrivacyType(99) // the maximum id in the privacy mixing range (0 - 99)

	PrivacyDestinationFirst = PrivacyType(100)
	PrivacyDestination      = PrivacyType(101)
	PrivacyDestinationLast  = PrivacyType(199)

	PrivacyOriginFirst = PrivacyType(200)
	PrivacyOrigin      = PrivacyType(201)
	PrivacyOriginLast  = PrivacyType(299)

	PrivacyCollateralCreationFirst = PrivacyType(300)
	PrivacyCollateralCreation      = PrivacyType(301)
	PrivacyCollateralCreationLast  = PrivacyType(399)

	PrivacyCollateralPaymentFirst = PrivacyType(400)
	PrivacyCollateralPayment      = PrivacyType(401)
	PrivacyCollateralPaymentLast  = PrivacyType(499)
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
