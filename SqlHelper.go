package goSqlHelper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type SqlHelper struct{
	Connection *sql.DB
	context context.Context
	tx *sql.Tx
	debugMod bool
}

const QUERY_BUFFER_SIZE=20

/**
@todo no sql

	var obj=new(tb1)
	con.Insert(obj)
	obj.setup(conn)
	obj.Select("id,val").Where("id=2").QueryList()
	sqlHelper.Select("id,val").Where("id=2").QueryList()

*/
func MysqlOpen(connectionStr string) (*SqlHelper,error){

	sqlHelper :=new (SqlHelper)
	err:= sqlHelper.Open(connectionStr)
	if err!=nil {
		return nil ,err
	}
	return sqlHelper,nil
}

func New(connectionStr string) (*SqlHelper,error){
	return MysqlOpen(connectionStr)
}

/**
begin context
*/
func (this *SqlHelper) BeginContext(ctx context.Context) *SqlHelperRunner{
	runner :=new(SqlHelperRunner)
	runner.SetDB(this.Connection)
	runner.SetContext(ctx)
	return runner
}

/**
begin a trasnaction
*/
func (this *SqlHelper) Begin() *SqlHelperRunner{
	runner :=new(SqlHelperRunner)
	runner.SetDB(this.Connection)
	runner.Begin()
	return runner
}

/**
print sql and parameter at prepare exeucting
 */
func (this *SqlHelper) OpenDebug(){
	this.debugMod=true
}

/**
begin a trasnaction
*/
func (this *SqlHelper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*SqlHelperRunner,error) {
	runner :=new(SqlHelperRunner)
	runner.SetDB(this.Connection)
	err:= runner.BeginTx(ctx,opts)
	if err!=nil {
		return nil,err
	}
	return runner,nil
}

/**
   open db
*/
func (this *SqlHelper) Open(connectionStr string) error{
	var err error
	
//	sql.Open
	this.Connection,err = sql.Open("mysql",connectionStr)
	if err!=nil {
		return errors.New(fmt.Sprintf("数据库链接失败:%s",err.Error()))
	}
	err=this.Connection.Ping();
	if err!=nil {
		return err
	}
	return nil
}

/**
set db object
*/
func (this *SqlHelper) SetDB (conn *sql.DB) {
		this.Connection=conn
}


/**
get Querying handler
*/
func (this *SqlHelper) Querying(sql string,args ...interface{})(*Querying,error){
	if this.debugMod {
		fmt.Println(sql)
		fmt.Println(args)
	}
	var rows ,err = this.query(sql,args...)
	if err!=nil {
		return nil, err
	}
	querying:= NewQuerying(rows)
	return querying,nil
}
/**
  read a int value
*/
func (this *SqlHelper) QueryScalarInt(sql string, args ...interface{})(int, error) {
	if this.debugMod {
		fmt.Println(sql)
		fmt.Println(args)
	}
	var rows ,err = this.query(sql,args...)
	if err!=nil {
		return 0,err
	}
	defer rows.Close()
	if rows.Next() {
		var val int
		err = rows.Scan(&val)
		return val,nil
	}
	return 0, NoFoundError
}

/*
execute sql
*/
func (this *SqlHelper) Exec(sql string,args ...interface{})(sql.Result,error){
	if this.debugMod {
		fmt.Println(sql)
		fmt.Println(args)
	}
	stmt,err:=this.prepare(sql)
	if err!=nil {
		return nil, err
	}
	defer stmt.Close()
	result,err := stmt.Exec(args...)
	if err!=nil {
		return nil, err
	}
	return result,nil
}

/*
execute insert sql
*/
func (this *SqlHelper) ExecInsert(sql string, args ...interface{})(int64,error){
	result,err := this.Exec(sql,args...)
	if err!=nil {
		return 0,err
	}
	
	id,err2 := result.LastInsertId()
	if err2 != nil {
		return 0, err2
	}
	return id , nil
}
/*
execute update or delete sql
*/
func (this *SqlHelper) ExecUpdateOrDel(sql string, args ...interface{})(int64,error){
	result,err := this.Exec(sql,args...)
	if err!=nil {
		return 0,err
	}
	
	cnt,err2 := result.RowsAffected()
	if err2 != nil {
		return 0, err2
	}
	return cnt , nil
}


/*
    close db pool
*/
func (this *SqlHelper) Close() error{
	err := this.Connection.Close()
	return err
}

// get auto sql
func(this *SqlHelper) Auto() *AutoSql{
	return NewAutoSql(this)
}