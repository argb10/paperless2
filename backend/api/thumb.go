package api

import (
	"io/ioutil"
	"net/http"

	"github.com/guad/paperless2/backend/api/user"
	"github.com/guad/paperless2/backend/crypto"

	"github.com/globalsign/mgo/bson"
	"github.com/guad/paperless2/backend/db"
	"github.com/guad/paperless2/backend/model"
	"github.com/guad/paperless2/backend/storage"
	"github.com/labstack/echo"
	"github.com/minio/minio-go"
)

func GetThumbnail(c echo.Context) error {
	doc := c.Param("doc")
	userid := user.GetUserID(c)

	sesh := db.Ctx()
	defer sesh.Close()
	col := sesh.DB("paperless").C("documents")

	if !bson.IsObjectIdHex(doc) {
		return c.JSON(http.StatusBadRequest, struct{}{})
	}

	id := bson.ObjectIdHex(doc)

	var document model.Document

	err := col.Find(bson.M{"_id": id, "user_id": bson.ObjectIdHex(userid)}).One(&document)

	if err != nil {
		return err
	}

	if document.UserID.Hex() != userid {
		return c.String(http.StatusForbidden, "")
	}

	if document.ThumbnailPath == "" {
		return c.String(http.StatusNotFound, "")
	}

	key := document.ThumbnailPath

	file, err := storage.S3.GetObject(
		storage.DocumentBucket,
		key,
		minio.GetObjectOptions{},
	)

	if err != nil {
		return err
	}

	defer file.Close()

	encryptedBytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	decrypted := crypto.Decrypt(encryptedBytes)

	return c.Blob(http.StatusOK, "image/png", decrypted)
}
