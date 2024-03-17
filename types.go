package main

import "time"

type DeleteActorReq struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UpdateActorReq struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Sex       string `json:"sex"`
	// DateOfBirth ??? `json:"dateOfBirth"`
}

type CreateActorReq struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Sex       string `json:"sex"`
	//DateOfBirth time.Time `json:"dateOfBirth"` dont provide this yet
}

type Actor struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Sex         string    `json:"sex"`
	DateOfBirth time.Time `json:"dateOfBirth"` //DateOfBirth fix the DOB or just make it a string
}

////

type DeleteMovieReq struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"releaseDate"`
}

type UpdateMovieReq struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"releaseDate"`
	Rating      byte      `json:"rating"`
	// Starring    []*Actor  `json:"starring"`

}

type CreateMovieReq struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"releaseDate"`
	Rating      byte      `json:"rating"`
	Starring    []*Actor  `json:"starring"`
}

type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"releaseDate"`
	Rating      byte      `json:"rating"`
	Starring    []*Actor  `json:"starring"`
}

// type DateOfBirth struct {
// 	time.Time
// }

// func (dob *DateOfBirth) UnmarshalJSON(b []byte) error {
// 	customLayout := "15-02-2003"

// 	parsedTime, err := time.Parse(`"`+customLayout+`"`, string(b))
// 	if err != nil {
// 		return err
// 	}

// 	dob.Time = parsedTime

// 	return nil
// }

func NewActor(firstName, lastName, sex string) *Actor {
	return &Actor{
		FirstName:   firstName,
		LastName:    lastName,
		Sex:         sex,
		DateOfBirth: time.Now().UTC(),
	}
}

func NewMovie(title, desc string, rating byte, starring []*Actor) *Movie {
	return &Movie{
		Title:       title,
		Description: desc,
		ReleaseDate: time.Now().UTC(),
		Rating:      rating,
		Starring:    starring,
	}
}
