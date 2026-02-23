package glclient

import "fmt"

var (
	ErrTokenRequired         = fmt.Errorf("token is required")
	ErrClientCreationFailed  = fmt.Errorf("failed to create client")
	ErrGetGroupFailed        = fmt.Errorf("failed to get group")
	ErrListGroupsFailed      = fmt.Errorf("failed to list groups")
	ErrCurrentUserFailed     = fmt.Errorf("failed to get current user")
	ErrListProjectsFailed    = fmt.Errorf("failed to list projects")
	ErrListIssuesFailed      = fmt.Errorf("failed to list issues")
	ErrListGroupIssuesFailed = fmt.Errorf("failed to list group issues")
)
