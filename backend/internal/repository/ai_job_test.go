package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

type jsonArgConnector struct{ value any }

func (c *jsonArgConnector) Connect(context.Context) (driver.Conn, error) {
	return &jsonArgConn{connector: c}, nil
}
func (c *jsonArgConnector) Driver() driver.Driver { return jsonArgDriver{connector: c} }

type jsonArgDriver struct{ connector *jsonArgConnector }

func (d jsonArgDriver) Open(string) (driver.Conn, error) {
	return &jsonArgConn{connector: d.connector}, nil
}

type jsonArgConn struct{ connector *jsonArgConnector }

func (*jsonArgConn) Prepare(string) (driver.Stmt, error) { return nil, fmt.Errorf("not supported") }
func (*jsonArgConn) Close() error                        { return nil }
func (*jsonArgConn) Begin() (driver.Tx, error)           { return jsonArgTx{}, nil }
func (c *jsonArgConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return jsonArgTx{}, nil
}
func (c *jsonArgConn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) > 1 {
		c.connector.value = args[1].Value
	}
	return driver.RowsAffected(1), nil
}

type jsonArgTx struct{}

func (jsonArgTx) Commit() error   { return nil }
func (jsonArgTx) Rollback() error { return nil }

func TestCompletePassesAIResultAsJSONText(t *testing.T) {
	connector := &jsonArgConnector{}
	db := sql.OpenDB(connector)
	defer db.Close()
	result := model.AIResult{Summary: "Site", ServiceCategory: "Website", Priority: "Low", PriorityReason: "Small scope", MissingInfo: []string{}}
	if err := (&briefRepository{db: db}).Complete(context.Background(), model.AIJob{ID: "job", BriefID: "brief"}, result, nil); err != nil {
		t.Fatal(err)
	}
	want := `{"summary":"Site","service_category":"Website","priority":"Low","priority_reason":"Small scope","missing_info":[]}`
	if got, ok := connector.value.(string); !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("JSONB argument = %T(%v), want string(%s)", connector.value, connector.value, want)
	}
}
