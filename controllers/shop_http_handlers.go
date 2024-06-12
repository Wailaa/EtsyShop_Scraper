package controllers

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Shop) CreateNewShopRequest(ctx *gin.Context) {

	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
	ShopRequest := &models.ShopRequest{}
	var shop NewShopRequest

	if err := ctx.ShouldBindJSON(&shop); err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, "failed to get the Shop's name", nil)
		return
	}

	ShopRequest.AccountID = currentUserUUID
	ShopRequest.ShopName = shop.ShopName

	existedShop, err := s.Operations.GetShopByName(shop.ShopName)
	if err != nil && err.Error() != "no Shop was Found ,error: record not found" {
		HandleResponse(ctx, err, http.StatusBadRequest, "internal error", nil)

		ShopRequest.Status = "failed"
		s.Operations.CreateShopRequest(ShopRequest)

		return

	} else if existedShop != nil {
		HandleResponse(ctx, nil, http.StatusBadRequest, "Shop already exists", nil)

		ShopRequest.Status = "denied"
		s.Operations.CreateShopRequest(ShopRequest)

		return
	}

	ShopRequest.Status = "Pending"
	s.Operations.CreateShopRequest(ShopRequest)

	HandleResponse(ctx, nil, http.StatusOK, "shop request received successfully", nil)

	go s.Operations.CreateNewShop(ShopRequest)

}

func (s *Shop) FollowShop(ctx *gin.Context) {

	var shopToFollow *FollowShopRequest
	if err := ctx.ShouldBindJSON(&shopToFollow); err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)

	requestedShop, err := s.Operations.GetShopByName(shopToFollow.FollowShopName)
	if err != nil {
		if err.Error() == "record not found" {
			HandleResponse(ctx, err, http.StatusBadRequest, "shop not found", nil)
			return
		}
		HandleResponse(ctx, err, http.StatusBadRequest, "error while processing the request", nil)
		return
	}
	if err := s.Operations.EstablishAccountShopRelation(requestedShop, currentUserUUID); err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
	}

	HandleResponse(ctx, err, http.StatusOK, "following shop", nil)

}

func (s *Shop) UnFollowShop(ctx *gin.Context) {

	var unFollowShop *UnFollowShopRequest
	if err := ctx.ShouldBindJSON(&unFollowShop); err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)

	requestedShop, err := s.Operations.GetShopByName(unFollowShop.UnFollowShopName)
	if err != nil {
		if err.Error() == "record not found" {
			HandleResponse(ctx, err, http.StatusBadRequest, "shop not found", nil)
			return
		}
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := s.Shop.UpdateAccountShopRelation(requestedShop, currentUserUUID); err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	HandleResponse(ctx, nil, http.StatusOK, "Unfollowed shop", nil)

}

func (s *Shop) HandleGetShopByID(ctx *gin.Context) {

	ShopID := ctx.Param("shopID")
	ShopIDToUint, err := utils.StringToUint(ShopID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, "failed to get Shop id", nil)
		return
	}
	Shop, err := s.Operations.GetShopByID(ShopIDToUint)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}
	HandleResponse(ctx, nil, http.StatusOK, "", Shop)

}

func (s *Shop) HandleGetItemsByShopID(ctx *gin.Context) {
	ShopID := ctx.Param("shopID")
	ShopIDToUint, err := utils.StringToUint(ShopID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, "failed to get Shop id", nil)
		return
	}
	Items, err := s.Operations.GetItemsByShopID(ShopIDToUint)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	HandleResponse(ctx, nil, http.StatusOK, "", Items)
}

func (s *Shop) HandleGetItemsCountByShopID(ctx *gin.Context) {
	ShopID := ctx.Param("shopID")
	ShopIDToUint, err := utils.StringToUint(ShopID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, "failed to get Shop id", nil)
		return
	}
	Items, err := s.GetItemsCountByShopID(ShopIDToUint)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	HandleResponse(ctx, nil, http.StatusOK, "", Items)
}

func (s *Shop) HandleGetSoldItemsByShopID(ctx *gin.Context) {
	ShopID := ctx.Param("shopID")
	ShopIDToUint, err := utils.StringToUint(ShopID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, "failed to get Shop id", nil)
		return
	}

	Items, err := s.Operations.GetSoldItemsByShopID(ShopIDToUint)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}
	HandleResponse(ctx, nil, http.StatusOK, "", Items)
}

func (s *Shop) ProcessStatsRequest(ctx *gin.Context) {

	ShopID := ctx.Param("shopID")
	Period := ctx.Param("period")
	ShopIDToUint, err := utils.StringToUint(ShopID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, "failed to get Shop id", nil)
		return
	}

	year, month, day := 0, 0, 0

	switch Period {
	case "lastSevenDays":
		day = -6
	case "lastThirtyDays":
		day = -29
	case "lastThreeMonths":
		month = -3
	case "lastSixMonths":
		month = -6
	case "lastYear":
		year = -1
	default:
		err := errors.New("invalid period provided")
		HandleResponse(ctx, err, http.StatusInternalServerError, err.Error(), nil)
		return

	}

	date := time.Now().AddDate(year, month, day)
	dateMidnight := utils.TruncateDate(date)

	LastSevenDays, err := s.Operations.GetSellingStatsByPeriod(ShopIDToUint, dateMidnight)
	if err != nil {
		HandleResponse(ctx, err, http.StatusInternalServerError, "error while handling stats", nil)
		return
	}

	HandleResponse(ctx, nil, http.StatusOK, "", gin.H{"stats": LastSevenDays})

}
