package twoway

import (
	context "context"
	driver "database/sql/driver"
	json "encoding/json"
	fmt "fmt"
	squirrel "github.com/Masterminds/squirrel"
	sqlx "github.com/jmoiron/sqlx"
	squirrel1 "github.com/roderm/gotools/squirrel"
)

var _ = fmt.Sprintf
var _ = context.TODO
var _ = driver.IsValue
var _ = json.Valid
var _ = squirrel.Select
var _ = sqlx.Connect
var _ = squirrel1.EqCall{}

type TwowayStore struct {
	conn *sqlx.DB
}

func NewTwowayStore(conn *sqlx.DB) *TwowayStore {
	return &TwowayStore{conn}
}

type MatchList map[interface{}]*Match

func (m *Match) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"id": m.MatchId,
	}
	return pk
}
func (m *Match) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Match' failed: %s", string(buff), err)
	}
	return nil
}

type queryMatchConfig struct {
	Store        *TwowayStore
	filter       []squirrel.Sqlizer
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

func MatchFilter(filter ...squirrel.Sqlizer) MatchOption {
	return func(config *queryMatchConfig) {
		config.filter = append(config.filter, filter...)
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
				squirrel1.EqCall{"id": func() interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}},
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
				squirrel1.EqCall{"id": func() interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}},
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

func (s *TwowayStore) Match(ctx context.Context, opts ...MatchOption) (MatchList, error) {
	config := &queryMatchConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(MatchList),
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
		_, err = s.Team(ctx, config.optsGuest...)
		if err != nil {
			return config.rows, err
		}
	}
	if config.loadHome {
		_, err = s.Team(ctx, config.optsHome...)
		if err != nil {
			return config.rows, err
		}
	}
	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *TwowayStore) GetMatchSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"id", "date", JSON_BUILD_OBJECT('id', "home") AS home, JSON_BUILD_OBJECT('id', "guest") AS guest, "home_score", "guest_score"`).
		From("\"tbl_match\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TwowayStore) selectMatch(ctx context.Context, config *queryMatchConfig, withRow func(*Match)) error {
	query := s.GetMatchSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectMatch': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Match{}
	for cursor.Next() {
		row := new(Match)
		err = cursor.StructScan(row)
		if err == nil {
			withRow(row)
			resultRows = append(resultRows, row)
		} else {
			return fmt.Errorf("sqlx.StructScan failed: %s", err)
		}

	}
	// err = sqlx.StructScan(cursor, &resultRows)
	// if err != nil {
	// 	return fmt.Errorf("sqlx.StructScan failed: %s", err)
	// }
	// for _, row := range resultRows {
	// 	withRow(row)
	// }
	return nil
}

type PlayerList map[interface{}]*Player

func (m *Player) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"id": m.PlayerId,
	}
	return pk
}
func (m *Player) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Player' failed: %s", string(buff), err)
	}
	return nil
}

type queryPlayerConfig struct {
	Store        *TwowayStore
	filter       []squirrel.Sqlizer
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

func PlayerFilter(filter ...squirrel.Sqlizer) PlayerOption {
	return func(config *queryPlayerConfig) {
		config.filter = append(config.filter, filter...)
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
				squirrel1.EqCall{"id": func() interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}},
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

func (s *TwowayStore) Player(ctx context.Context, opts ...PlayerOption) (PlayerList, error) {
	config := &queryPlayerConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(PlayerList),
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
		_, err = s.Team(ctx, config.optsTeam...)
		if err != nil {
			return config.rows, err
		}
	}
	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *TwowayStore) GetPlayerSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"id", "name", "number", JSON_BUILD_OBJECT('id', "team") AS team`).
		From("\"tbl_player\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TwowayStore) selectPlayer(ctx context.Context, config *queryPlayerConfig, withRow func(*Player)) error {
	query := s.GetPlayerSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectPlayer': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Player{}
	for cursor.Next() {
		row := new(Player)
		err = cursor.StructScan(row)
		if err == nil {
			withRow(row)
			resultRows = append(resultRows, row)
		} else {
			return fmt.Errorf("sqlx.StructScan failed: %s", err)
		}

	}
	// err = sqlx.StructScan(cursor, &resultRows)
	// if err != nil {
	// 	return fmt.Errorf("sqlx.StructScan failed: %s", err)
	// }
	// for _, row := range resultRows {
	// 	withRow(row)
	// }
	return nil
}

type TeamList map[interface{}]*Team

func (m *Team) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"id": m.TeamId,
	}
	return pk
}
func (m *Team) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Team' failed: %s", string(buff), err)
	}
	return nil
}

type queryTeamConfig struct {
	Store        *TwowayStore
	filter       []squirrel.Sqlizer
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

func TeamFilter(filter ...squirrel.Sqlizer) TeamOption {
	return func(config *queryTeamConfig) {
		config.filter = append(config.filter, filter...)
	}
}

func TeamOnRow(cb func(*Team)) TeamOption {
	return func(s *queryTeamConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (l TeamList) LoadPlayers(s *TwowayStore, ctx context.Context, opts ...PlayerOption) error {
	ids := []interface{}{}
	for p := range l {
		ids = append(ids, p)
	}
	opts = append(opts,
		PlayerOnRow(func(row *Player) {
			parent_id := row.PlayerId
			if _, ok := l[parent_id]; ok {
				l[parent_id].Players = append(l[parent_id].Players, row)
			}
		}),
		PlayerFilter(
			squirrel.Eq{"id": ids},
		),
	)
	_, err := s.Player(ctx, opts...)
	return err
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
				squirrel1.EqCall{"id": func() interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}},
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

func (s *TwowayStore) Team(ctx context.Context, opts ...TeamOption) (TeamList, error) {
	config := &queryTeamConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(TeamList),
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
		_, err = s.Player(ctx, config.optsPlayers...)
		if err != nil {
			return config.rows, err
		}
	}
	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *TwowayStore) GetTeamSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"id", "league"`).
		From("\"tbl_team\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TwowayStore) selectTeam(ctx context.Context, config *queryTeamConfig, withRow func(*Team)) error {
	query := s.GetTeamSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectTeam': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Team{}
	for cursor.Next() {
		row := new(Team)
		err = cursor.StructScan(row)
		if err == nil {
			withRow(row)
			resultRows = append(resultRows, row)
		} else {
			return fmt.Errorf("sqlx.StructScan failed: %s", err)
		}

	}
	// err = sqlx.StructScan(cursor, &resultRows)
	// if err != nil {
	// 	return fmt.Errorf("sqlx.StructScan failed: %s", err)
	// }
	// for _, row := range resultRows {
	// 	withRow(row)
	// }
	return nil
}
