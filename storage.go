package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateActor(*Actor) error
	UpdateActor(*UpdateActorReq) error
	GetActors() ([]*Actor, error)
	DeleteActor(int, string, string) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStore, error) {
	connStr := "user=postgres port=5433 dbname=postgres password=root sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("connected to db")
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	s.createActorTable()
	s.createMovieTable()
	return nil
}

func (s *PostgresStore) createActorTable() error {
	query := `create table if not exists actor (
		id serial primary key,
		first_name varchar(50) not null,
		last_name varchar(50) not null,
		sex varchar(50) not null,
		date_of_birth timestamp
	)`
	_, err := s.db.Exec(query)

	return err

}

func (s *PostgresStore) createMovieTable() error {
	query := `create table if not exists movie (
		id serial primary key,
		title varchar(150) not null,
		description varchar(1000),
		release_date timestamp not null,
		date_of_birth timestamp
		starring varchar[]
	)`
	_, err := s.db.Exec(query)

	return err

}

func (s *PostgresStore) CreateActor(act *Actor) error {
	query := `insert into actor 
	(first_name,last_name,sex,date_of_birth)
	values ($1,$2,$3,$4)`
	resp, err := s.db.Query(query, act.FirstName, act.LastName, act.Sex, act.DateOfBirth)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) UpdateActor(updateData *UpdateActorReq) error {
	query := "UPDATE actor SET "
	var params []interface{}
	var setFields []string
	paramIndex := 1
	if updateData.FirstName != "" {
		setFields = append(setFields, fmt.Sprintf("first_name = $%d", paramIndex))
		params = append(params, updateData.FirstName)
		paramIndex++
	}
	if updateData.LastName != "" {
		setFields = append(setFields, fmt.Sprintf("last_name = $%d", paramIndex))
		params = append(params, updateData.LastName)
		paramIndex++
	}
	if updateData.Sex != "" {
		setFields = append(setFields, fmt.Sprintf("sex = $%d", paramIndex))
		params = append(params, updateData.Sex)
		paramIndex++
	}
	// more fields here

	query += strings.Join(setFields, ", ")
	query += " WHERE id = $"
	query += fmt.Sprint(paramIndex)

	params = append(params, updateData.ID)

	_, err := s.db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("failed to update actor")
	}
	return nil
}

func (s *PostgresStore) DeleteActor(id int, firstName, lastName string) error {
	_, err := s.db.Query(`delete from actor where (id = $1)`, id)
	/* and first_name = $2 and last_name = $3) or (id = $1 and first_name = $2)
	or (id = $1 and last_name = $3)
	or id = $1 */ // fix later of never
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetActors() ([]*Actor, error) {
	rows, err := s.db.Query("select * from actor")
	if err != nil {
		return nil, err
	}
	actors := []*Actor{}
	for rows.Next() {
		actor, err := scanIntoActor(rows)
		if err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}
	return actors, nil

}

func scanIntoActor(rows *sql.Rows) (*Actor, error) {
	actor := new(Actor)
	err := rows.Scan(
		&actor.ID,
		&actor.FirstName,
		&actor.LastName,
		&actor.Sex,
		&actor.DateOfBirth)
	return actor, err
}
