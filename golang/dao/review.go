package dao

import (
	_"log"
	_"reflect"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/syunsukeA/oreno_ramen/golang/domain/object"

	"github.com/jmoiron/sqlx"
)

// repsitory.Reviewのimplements
type Review struct {
	DB *sqlx.DB
}

func (r *Review)GetUnvisitedReviews() (ROs []*object.Review){
	entity := new(object.Review)
	ROs = append(ROs, entity)
	return ROs
}

func (r *Review)AddReviewAndShop(c *gin.Context, shopID string, userID int64, shopname string, req *object.CreateReviewRequest) (ro *object.Review, err error) {
	ro = new(object.Review)
	// トランザクション処理の開始
	tx, err := r.DB.BeginTxx(c, nil)
	if err != nil {
        return nil, err
    }
	defer func() (*object.Review, error) {
        switch r := recover(); {
		case r != nil:
            tx.Rollback()
            return nil, err
        case err != nil:
            tx.Rollback()
			return nil, err
		default:
			return ro, nil
		}
    }()
	so := new(object.Shop)
	// shopに該当データがあるか確認
	q := `SELECT * from shops WHERE shop_id = ?`
	err = tx.QueryRowxContext(c, q, shopID).StructScan(so)
	if err != nil {
		// shopになかったらshopデータの追加
		if err == sql.ErrNoRows {
			q = `INSERT INTO shops (shop_id, user_id, shopname) VALUES (?, ?, ?)`
			res, err := tx.Exec(q, shopID, userID, shopname)
			if err != nil {
				return nil, err
			}
			nrows, err := res.RowsAffected()
			if err != nil {
				return nil, err
			}
			if nrows <= 0 {
				return nil, sql.ErrNoRows
			}
		} else { // その他のerrorはシンプルに誤作動なのでerrをreturn
			return nil, err
		}
	}
	// reviewデータの追加
	// ToDo: 構造体駆使して短くする？
	q = `INSERT INTO reviews (user_id, shop_id, shopname, dishname, content, evaluate, review_img) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(q, userID, shopID, shopname, req.DishName, req.Content, req.Evaluate, req.ReviewImg)
		if err != nil {
			return nil, err
		}
		// ToDo: 作成したreviewをResBodyに入れるために検索？？
		// review_id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		// ToDo: もしかしたらこの辺の処理いらないかも？
		nrows, err := res.RowsAffected()
		if err != nil {
			return nil, err
		}
		if nrows <= 0 {
			return nil, sql.ErrNoRows
		}
	// トランザクション処理をコミット
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return ro, nil
}