package root

import (
	"fmt"

	"github.com/chazzychouse/g2o/internal/styles"
)

func Run() error {
	fmt.Println(styles.Banner.Render("g2o") + " — Use --help to see available commands.")
	return nil
}
