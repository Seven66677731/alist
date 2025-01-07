package handles

import (
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/strm"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"path"
)

func GenerateStrm(c *gin.Context) {
	var req ListReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	user := c.MustGet("user").(*model.User)
	reqPath, err := user.JoinPath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	meta, err := op.GetNearestMeta(reqPath)
	if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
		common.ErrorResp(c, err, 500, true)
		return
	}
	c.Set("meta", meta)

	if !common.CanAccess(user, meta, reqPath, req.Password) {
		common.ErrorStrResp(c, "password is incorrect or you have no permission", 403)
		return
	}
	if !user.CanWrite() && !common.CanWrite(meta, reqPath) && req.Refresh {
		common.ErrorStrResp(c, "Refresh without permission", 403)
		return
	}

	filePaths, err := listAllFile(c, reqPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	resMap := strm.WriteFiles(filePaths)

	deleteFiles := strm.DeleteExtraFiles(reqPath, filePaths)
	resMap["deleteFiles"] = deleteFiles
	common.SuccessResp(c, resMap)
}

func listAllFile(c *gin.Context, parentDir string) ([]string, error) {
	utils.Log.Debugf("Listing files in directory: %s", parentDir)
	var filePaths []string

	objs, err := fs.List(c, parentDir, &fs.ListArgs{Refresh: true})
	if err != nil {
		utils.Log.Errorf("Failed to list files in directory: %s, error: %v", parentDir, err)
		return nil, err
	}

	for i := range objs {
		obj := objs[i]
		if obj.IsDir() {
			subDirFiles, err := listAllFile(c, path.Join(parentDir, obj.GetName()))
			if err != nil {
				return nil, err
			}
			filePaths = append(filePaths, subDirFiles...)
		} else {
			filePaths = append(filePaths, path.Join(parentDir, obj.GetName()))
		}
	}
	return filePaths, nil
}
