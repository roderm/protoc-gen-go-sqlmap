package twoway

import (
	context "context"
	sql "database/sql"
	driver "database/sql/driver"
	json "encoding/json"
	fmt "fmt"
	pg "github.com/roderm/protoc-gen-go-sqlmap/lib/go/pg"
)

var _ = fmt.Sprintf
var _ = context.TODO
var _ = pg.NONE
var _ = sql.Open
var _ = driver.IsValue
var _ = json.Valid

type TwowayStore struct {
	conn *sql.DB
}

func NewTwowayStore(conn *sql.DB) *TwowayStore {
	return &TwowayStore{conn}
}

type queryMatchConfig struct {
	Store        *TwowayStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Match) error
	cb           []func(*Match)
	rows         map[interface{}]*Match
	loadHome     bool
	optsHome     []TeamOption
	loadGuest    bool
	optsGuest    []TeamOption
}

type MatchOption func(*queryMatchConfig)

func MatchPaging(page, length int) MatchOption {
	return func(config *queryMatchConfig) {
		config.start = length * page
		config.limit = length
	}
}
func MatchFilter(filter pg.Where) MatchOption {
	return func(config *queryMatchConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func MatchOnRow(cb func(*Match)) MatchOption {
	return func(s *queryMatchConfig) {
		s.cb = append(s.cb, cb)
	}
}

func MatchWithHome(opts ...TeamOption) MatchOption {
	return func(config *queryMatchConfig) {
		config.loadHome = true
		parent := make(map[interface{}][]*Match)
		config.cb = append(config.cb, func(row *Match) {
			child_key := row.GetHome().TeamId
			parent[child_key] = append(parent[child_key], row)
		})
		config.optsHome = append(opts,
			TeamFilter(
				pg.INCallabel("id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			TeamOnRow(func(row *Team) {
				children := parent[row.TeamId]
				for _, c := range children {
					c.Home = row
				}
			}),
		)
	}
}

func MatchWithGuest(opts ...TeamOption) MatchOption {
	return func(config *queryMatchConfig) {
		config.loadGuest = true
		parent := make(map[interface{}][]*Match)
		config.cb = append(config.cb, func(row *Match) {
			child_key := row.GetGuest().TeamId
			parent[child_key] = append(parent[child_key], row)
		})
		config.optsGuest = append(opts,
			TeamFilter(
				pg.INCallabel("id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			TeamOnRow(func(row *Team) {
				children := parent[row.TeamId]
				for _, c := range children {
					c.Guest = row
				}
			}),
		)
	}
}

func (s *TwowayStore) Match(ctx context.Context, opts ...MatchOption) (map[interface{}]*Match, error) {
	config := &queryMatchConfig{
		Store:  s,
		filter: pg.NONE(),
		limit:  1000,
		rows:   make(map[interface{}]*Match),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectMatch(ctx, config, func(row *Match) {
		config.rows[row.MatchId] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}

	if config.loadGuest {
		// github.com/roderm/protoc-gen-go-sqlmap/test/twoway/twoway

		_, err = s.Team(ctx, config.optsGuest...)

	}
	if err != nil {
		return config.rows, err
	}

	if config.loadHome {
		// github.com/roderm/protoc-gen-go-sqlmap/test/twoway/twoway

		_, err = s.Team(ctx, config.optsHome...)

	}
	if err != nil {
		return config.rows, err
	}

	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *TwowayStore) GetMatchSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "id", "date", "home", "guest", "home_score", "guest_score"
		FROM "tbl_match"
		%s`, where)

	if limit > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nLIMIT $%d", base)
		vals = append(vals, limit)
	}
	if start > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nOFFSET $%d", base)
		vals = append(vals, start)
	}
	return tpl, vals
}

func (s *TwowayStore) selectMatch(ctx context.Context, config *queryMatchConfig, withRow func(*Match)) error {
	query, vals := s.GetMatchSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectMatch': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectMatch' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Match{
			Home:  new(Team),
			Guest: new(Team),
		}
		err := cursor.Scan(&row.MatchId, &row.Date, &row.HomeScore, &row.GuestScore, &row.GetHome().TeamId, &row.GetGuest().TeamId)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}

type queryPlayerConfig struct {
	Store        *TwowayStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Player) error
	cb           []func(*Player)
	rows         map[interface{}]*Player
	loadTeam     bool
	optsTeam     []TeamOption
}

type PlayerOption func(*queryPlayerConfig)

func PlayerPaging(page, length int) PlayerOption {
	return func(config *queryPlayerConfig) {
		config.start = length * page
		config.limit = length
	}
}
func PlayerFilter(filter pg.Where) PlayerOption {
	return func(config *queryPlayerConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func PlayerOnRow(cb func(*Player)) PlayerOption {
	return func(s *queryPlayerConfig) {
		s.cb = append(s.cb, cb)
	}
}

func PlayerWithTeam(opts ...TeamOption) PlayerOption {
	return func(config *queryPlayerConfig) {
		config.loadTeam = true
		parent := make(map[interface{}][]*Player)
		config.cb = append(config.cb, func(row *Player) {
			child_key := row.GetTeam().TeamId
			parent[child_key] = append(parent[child_key], row)
		})
		config.optsTeam = append(opts,
			TeamFilter(
				pg.INCallabel("id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			TeamOnRow(func(row *Team) {
				children := parent[row.TeamId]
				for _, c := range children {
					c.Team = row
				}
			}),
		)
	}
}

func (s *TwowayStore) Player(ctx context.Context, opts ...PlayerOption) (map[interface{}]*Player, error) {
	config := &queryPlayerConfig{
		Store:  s,
		filter: pg.NONE(),
		limit:  1000,
		rows:   make(map[interface{}]*Player),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectPlayer(ctx, config, func(row *Player) {
		config.rows[row.PlayerId] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}

	if config.loadTeam {
		// github.com/roderm/protoc-gen-go-sqlmap/test/twoway/twoway

		_, err = s.Team(ctx, config.optsTeam...)

	}
	if err != nil {
		return config.rows, err
	}

	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *TwowayStore) GetPlayerSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "id", "name", "number", "team"
		FROM "tbl_player"
		%s`, where)

	if limit > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nLIMIT $%d", base)
		vals = append(vals, limit)
	}
	if start > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nOFFSET $%d", base)
		vals = append(vals, start)
	}
	return tpl, vals
}

func (s *TwowayStore) selectPlayer(ctx context.Context, config *queryPlayerConfig, withRow func(*Player)) error {
	query, vals := s.GetPlayerSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectPlayer': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectPlayer' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Player{
			Team: new(Team),
		}
		err := cursor.Scan(&row.PlayerId, &row.Name, &row.Number, &row.GetTeam().TeamId)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}

type queryTeamConfig struct {
	Store        *TwowayStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Team) error
	cb           []func(*Team)
	rows         map[interface{}]*Team
	loadPlayers  bool
	optsPlayers  []PlayerOption
}

type TeamOption func(*queryTeamConfig)

func TeamPaging(page, length int) TeamOption {
	return func(config *queryTeamConfig) {
		config.start = length * page
		config.limit = length
	}
}
func TeamFilter(filter pg.Where) TeamOption {
	return func(config *queryTeamConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func TeamOnRow(cb func(*Team)) TeamOption {
	return func(s *queryTeamConfig) {
		s.cb = append(s.cb, cb)
	}
}

func TeamWithPlayers(opts ...PlayerOption) TeamOption {
	return func(config *queryTeamConfig) {
		config.loadPlayers = true
		parent := make(map[interface{}]*Team)
		config.cb = append(config.cb, func(row *Team) {
			child_key := row.TeamId
			parent[child_key] = row
		})
		config.optsPlayers = append(opts,
			PlayerFilter(
				pg.INCallabel("id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			PlayerOnRow(func(row *Player) {
				parent_id := row.PlayerId
				if _, ok := parent[parent_id]; ok {
					parent[parent_id].Players = append(parent[parent_id].Players, row)
				}
			}),
		)
	}
}

func (s *TwowayStore) Team(ctx context.Context, opts ...TeamOption) (map[interface{}]*Team, error) {
	config := &queryTeamConfig{
		Store:  s,
		filter: pg.NONE(),
		limit:  1000,
		rows:   make(map[interface{}]*Team),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectTeam(ctx, config, func(row *Team) {
		config.rows[row.TeamId] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}

	if config.loadPlayers {
		// github.com/roderm/protoc-gen-go-sqlmap/test/twoway/twoway

		_, err = s.Player(ctx, config.optsPlayers...)

	}
	if err != nil {
		return config.rows, err
	}

	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *TwowayStore) GetTeamSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "id", "league"
		FROM "tbl_team"
		%s`, where)

	if limit > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nLIMIT $%d", base)
		vals = append(vals, limit)
	}
	if start > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nOFFSET $%d", base)
		vals = append(vals, start)
	}
	return tpl, vals
}

func (s *TwowayStore) selectTeam(ctx context.Context, config *queryTeamConfig, withRow func(*Team)) error {
	query, vals := s.GetTeamSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectTeam': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectTeam' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Team{
			Players: []*Player{},
		}
		err := cursor.Scan(&row.TeamId, &row.League)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
