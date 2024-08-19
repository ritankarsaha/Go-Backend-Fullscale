package controllers

import (
	"context"
	"fmt"
	"github.com/ritankarsaha/Go-Backend-Fullscale/database"
	"github.com/ritankarsaha/Go-Backend-Fullscale/helpers"
	"github.com/ritankarsaha/Go-Backend-Fullscale/models"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

