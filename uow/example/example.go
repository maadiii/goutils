package example

import (
	"context"

	"github.com/maadiii/utils/uow"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Run() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	repoFactory := NewRepoFactory(db)
	uow := uow.NewGorm(db, NewRepoFactory)

	service := service{
		entities:   repoFactory.Entity(),
		uowFactory: uow,
	}

	service.Create(context.Background())
}
