package sm

import (
	"errors"
)

var ErrNotFound = errors.New("Whatever you are looking for, it doesn't exist")

const (
	EasyTVApiKeyHeader  = "X-Easytv-Key"
	EasyTVSessionHeader = "X-Easytv-Session"

	SessionExpiration = 60 * 20 // 20 minutes

	RoleAdmin        = 0
	RoleContentOwner = 1

	OK = 200

	// Generic errors
	CodeMissingInput        = -400
	CodeNoSession           = -401
	CodeNotFound            = -404
	CodeInternalServerError = -500

	// Domain errors
	CodeTaskNotDisabled                    = -1
	CodeTaskHasActiveJobs                  = -2
	CodeJobAlreadyCanceled                 = -3
	CodeJobAlreadyCompleted                = -4
	CodeEmptyAsset                         = -5
	CodeTaskAlreadyExists                  = -8
	CodeTaskNoInputParameter               = -9
	CodeTaskNoOutputParameter              = -10
	CodeInvalidStartUrl                    = -11
	CodeInvalidCancelUrl                   = -12
	CodeInvalidInput                       = -13
	CodeInvalidPublicationdate             = -14
	CodeInvalidExpirationDate              = -15
	CodeJobStatusNotUpdatable              = -16
	CodeForbiddenAsset                     = -17
	CodeInvalidOutput                      = -18
	CodeNotCompletable                     = -19
	CodeLinkedParameterNotTheSameType      = -20
	CodeLinkedOutputNotFound               = -21
	CodeServiceNameInUse                   = -22
	CodeContentOwnerNameExists             = -23
	CodeContentOwnerUsernameExists         = -24
	CodeContentOwnerEmailExists            = -25
	CodeJobWithDisabledTasks               = -26
	CodePasswordIsTooShort                 = -27
	CodeInvalidCredentials                 = -28
	CodeNewPasswordDoesntMatchVerification = -29
)
