package internal

// Pro module third-party package imports
// These imports ensure that the special third-party packages used in the pro module
// are included in the build, even if they are not directly used in the main codebase.

import (
	_ "embed"
	// Web framework
	_ "github.com/GoYoko/web"

	// Wire dependency injection
	_ "github.com/google/wire"

	// UUID generation
	_ "github.com/google/uuid"

	// Echo web framework
	_ "github.com/labstack/echo/v4"
	// Embed directive for certificates
)
