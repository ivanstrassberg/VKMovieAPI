package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateActor(*Actor) error
	UpdateActor(*UpdateActorReq) error
	GetActors() ([]*Actor, error)
	DeleteActor(int) error
	GetActorById(int) (*Actor, error)
	DeleteActorData(*UpdateActorReq) error

	CreateMovie(*Movie) error
	SearchMovie(string) ([]*Movie, error)
	UpdateMovie(*UpdateMovieReq) error
	GetSortedMovies(string, string) ([]*Movie, error)
	DeleteMovie(int) error
	DeleteMovieData(*UpdateMovieReq) error
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
		first_name varchar(50),
		last_name varchar(50),
		sex varchar(50),
		date_of_birth timestamp,
		starring_in int[]
	)`
	_, err := s.db.Exec(query)

	return err

}

func (s *PostgresStore) createMovieTable() error {
	query := `CREATE TABLE IF NOT EXISTS movie (
		id SERIAL PRIMARY KEY,
		title VARCHAR(150) CHECK (LENGTH(title) >= 1 AND LENGTH(title) <= 150),
		description VARCHAR(1000) CHECK (LENGTH(description) <= 1000),
		release_date TIMESTAMP,
		rating SMALLINT CHECK (rating >= 0 AND rating <= 10),
		starring INT[]
	)`
	_, err := s.db.Exec(query)

	return err

}

func (s *PostgresStore) CreateActor(act *Actor) error {

	for _, movieID := range act.StarringIn {
		actorExists, err := s.movieExists(movieID)
		if err != nil {
			return err
		}
		if !actorExists {
			return fmt.Errorf("movie with ID %d does not exist", movieID)
		}
	}

	query := `insert into actor 
	(first_name,last_name,sex,date_of_birth,starring_in)
	values ($1,$2,$3,$4,$5)`
	resp, err := s.db.Query(query, act.FirstName, act.LastName, act.Sex, act.DateOfBirth, intSliceToArrayLiteral(act.StarringIn))
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) UpdateActor(updateData *UpdateActorReq) error {

	for _, movieID := range updateData.StarringIn {
		actorExists, err := s.movieExists(movieID)
		if err != nil {
			return err
		}
		if !actorExists {
			return fmt.Errorf("movie with ID %d does not exist", movieID)
		}
	}

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
	if len(updateData.StarringIn) != 0 {
		setFields = append(setFields, fmt.Sprintf("starring_in = $%d", paramIndex))
		params = append(params, intSliceToArrayLiteral(updateData.StarringIn))
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

func (s *PostgresStore) DeleteActor(id int) error {
	_, err := s.db.Query(`delete from actor where (id = $1)`, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) DeleteActorData(updateData *UpdateActorReq) error {

	query := "UPDATE actor SET "
	var params []interface{}
	var setFields []string
	paramIndex := 1
	if updateData.FirstName == "" {
		setFields = append(setFields, fmt.Sprintf("first_name = $%d", paramIndex))
		params = append(params, updateData.FirstName)
		paramIndex++
	}
	if updateData.LastName == "" {
		setFields = append(setFields, fmt.Sprintf("last_name = $%d", paramIndex))
		params = append(params, updateData.LastName)
		paramIndex++
	}
	if updateData.Sex == "" {
		setFields = append(setFields, fmt.Sprintf("sex = $%d", paramIndex))
		params = append(params, updateData.Sex)
		paramIndex++
	}
	if len(updateData.StarringIn) == 0 {
		setFields = append(setFields, fmt.Sprintf("starring_in = $%d", paramIndex))
		params = append(params, intSliceToArrayLiteral(updateData.StarringIn))
		paramIndex++
	}
	// more fields here

	query += strings.Join(setFields, ", ")
	query += " WHERE id = $"
	query += fmt.Sprint(paramIndex)

	params = append(params, updateData.ID)

	_, err := s.db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("failed to delete actor data")
	}
	return nil
}

func (s *PostgresStore) GetActors() ([]*Actor, error) {
	rows, err := s.db.Query("SELECT * FROM actor")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actors := []*Actor{}
	for rows.Next() {
		actor, err := scanIntoActor(rows)
		if err != nil {
			return nil, err
		}

		for _, movieID := range actor.StarringIn {
			movie, err := s.GetMovieByID(movieID)
			if err != nil {
				return nil, err
			}
			actor.StarringInDetails = append(actor.StarringInDetails, movie)
		}

		actors = append(actors, actor)
	}
	return actors, nil
}

func (s *PostgresStore) GetMovieByID(id int) (*Movie, error) {
	rows, err := s.db.Query("SELECT * FROM movie WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		movie, err := scanIntoMovie(rows)
		if err != nil {
			return nil, err
		}
		return movie, nil
	}
	return nil, fmt.Errorf("movie with ID %d not found", id)
}

func scanIntoActor(rows *sql.Rows) (*Actor, error) {
	actor := new(Actor)
	var starringIn pq.Int64Array
	err := rows.Scan(
		&actor.ID,
		&actor.FirstName,
		&actor.LastName,
		&actor.Sex,
		&actor.DateOfBirth,
		&starringIn)
	if err != nil {
		return nil, err
	}
	actor.StarringIn = make([]int, len(starringIn))
	for i, v := range starringIn {
		actor.StarringIn[i] = int(v)
	}

	return actor, err
}

func scanIntoMovie(rows *sql.Rows) (*Movie, error) {
	movie := new(Movie)
	var starring pq.Int64Array
	err := rows.Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.ReleaseDate,
		&movie.Rating,
		&starring)
	if err != nil {
		return nil, err
	}

	movie.Starring = make([]int, len(starring))
	for i, v := range starring {
		movie.Starring[i] = int(v)
	}

	return movie, nil
}

func (s *PostgresStore) SearchMovie(searchQuery string) ([]*Movie, error) {
	// Split the search query into words
	searchWords := strings.Fields(searchQuery)

	// Construct the SQL query with LIKE clauses for movie title and actor names
	query := `
        SELECT DISTINCT m.*
        FROM movie m
        LEFT JOIN actor a ON m.starring @> ARRAY[a.id]
        WHERE `

	var queryParams []interface{}
	for i, word := range searchWords {
		if i > 0 {
			query += " AND "
		}
		query += "(m.title ILIKE '%' || $" + strconv.Itoa(i*3+1)
		query += " || '%' OR a.first_name ILIKE '%' || $" + strconv.Itoa(i*3+2)
		query += " || '%' OR a.last_name ILIKE '%' || $" + strconv.Itoa(i*3+3) + ")"
		queryParams = append(queryParams, word, word, word)
	}

	// Execute the query with parameters
	rows, err := s.db.Query(query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the result rows into Movie objects
	var movies []*Movie
	for rows.Next() {
		movie, err := scanIntoMovie(rows)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (s *PostgresStore) GetSortedMovies(keyWordSortParam, keyWord string) ([]*Movie, error) {
	var query string
	if keyWordSortParam != " " && keyWord != " " {

		query = fmt.Sprintf("select * from movie ORDER BY %s %s;", keyWordSortParam, keyWord)
	} else {
		query = "select * from movie ORDER BY rating DESC"
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []*Movie{}
	for rows.Next() {
		movie, err := scanIntoMovie(rows)
		if err != nil {
			return nil, err
		}

		for _, actorID := range movie.Starring {
			actor, err := s.GetActorById(actorID)
			if err != nil {
				return nil, err
			}
			movie.StarringDetails = append(movie.StarringDetails, actor)
		}

		movies = append(movies, movie)
	}
	return movies, nil
}

func (s *PostgresStore) CreateMovie(movie *Movie) error {
	query := `insert into movie 
	(title,description,release_date,rating,starring)
	values ($1,$2,$3,$4,$5)`
	resp, err := s.db.Query(query, movie.Title, movie.Description, movie.ReleaseDate, movie.Rating, intSliceToArrayLiteral(movie.Starring))
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) UpdateMovie(updateData *UpdateMovieReq) error {

	for _, actorID := range updateData.Starring {
		actorExists, err := s.actorExists(actorID)
		if err != nil {
			return err
		}
		if !actorExists {
			return fmt.Errorf("movie with ID %d does not exist", actorID)
		}
	}

	query := "UPDATE movie SET "
	var params []interface{}
	var setFields []string
	paramIndex := 1
	if len(updateData.Title) != 0 {
		setFields = append(setFields, fmt.Sprintf("title = $%d", paramIndex))
		params = append(params, updateData.Title)
		paramIndex++
	}
	if len(updateData.Description) != 0 {
		setFields = append(setFields, fmt.Sprintf("description = $%d", paramIndex))
		params = append(params, updateData.Description)
		paramIndex++
	}
	if updateData.ReleaseDate.IsZero() {
		setFields = append(setFields, fmt.Sprintf("release_date = $%d", paramIndex))
		params = append(params, updateData.ReleaseDate)
		paramIndex++
	}
	if updateData.Rating >= 0 {
		setFields = append(setFields, fmt.Sprintf("rating = $%d", paramIndex))
		params = append(params, updateData.Rating)
		paramIndex++
	}
	if len(updateData.Starring) != 0 {
		setFields = append(setFields, fmt.Sprintf("starring = $%d", paramIndex))
		params = append(params, intSliceToArrayLiteral(updateData.Starring))
		paramIndex++
	}
	// more fields here

	query += strings.Join(setFields, ", ")
	query += " WHERE id = $"
	query += fmt.Sprint(paramIndex)

	params = append(params, updateData.ID)

	_, err := s.db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("failed to update movie")
	}
	return nil
}

func (s *PostgresStore) DeleteMovieData(updateData *UpdateMovieReq) error {

	query := "UPDATE movie SET "
	var params []interface{}
	var setFields []string
	paramIndex := 1
	if updateData.Title == "" {
		setFields = append(setFields, fmt.Sprintf("title = $%d", paramIndex))
		params = append(params, updateData.Title)
		paramIndex++
	}
	if updateData.Description == "" {
		setFields = append(setFields, fmt.Sprintf("description = $%d", paramIndex))
		params = append(params, updateData.Description)
		paramIndex++
	}
	if updateData.ReleaseDate.IsZero() {
		setFields = append(setFields, fmt.Sprintf("release_date = $%d", paramIndex))
		params = append(params, updateData.ReleaseDate)
		paramIndex++
	}
	if updateData.Rating == 0 {
		setFields = append(setFields, fmt.Sprintf("rating = $%d", paramIndex))
		params = append(params, updateData.Rating)
		paramIndex++
	}
	if len(updateData.Starring) == 0 {
		setFields = append(setFields, fmt.Sprintf("starring = $%d", paramIndex))
		params = append(params, intSliceToArrayLiteral(updateData.Starring))
		paramIndex++
	}
	// more fields here

	query += strings.Join(setFields, ", ")
	query += " WHERE id = $"
	query += fmt.Sprint(paramIndex)

	params = append(params, updateData.ID)

	_, err := s.db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("failed to delete movie data")
	}
	return nil
}

func (s *PostgresStore) DeleteMovie(id int) error {
	_, err := s.db.Query(`delete from movie where (id = $1)`, id)
	if err != nil {
		return err
	}
	return nil
}

func intSliceToArrayLiteral(slice []int) string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, v := range slice {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Itoa(v))
	}
	sb.WriteByte('}')
	return sb.String()
}

func (s *PostgresStore) GetActorById(id int) (*Actor, error) {

	rows, err := s.db.Query("select	* from actor where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoActor(rows)
	}
	return nil, fmt.Errorf("actor %d not found", id)
}

func (s *PostgresStore) actorExists(actorID int) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM actor WHERE id = $1", actorID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error checking actor existence: %w", err)
	}
	return count > 0, nil
}

func (s *PostgresStore) movieExists(movieID int) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM movie WHERE id = $1", movieID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error checking movie existence: %w", err)
	}
	return count > 0, nil
}
