package search

import (
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/service/cloudsearchdomain"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/goproducts/config"
	"github.com/goproducts/database"
	"github.com/goproducts/dto"
	"github.com/goproducts/errors"
)

// ProductScan scans product documents for matches
func ProductScan(svc *database.DB, searchTerm string) []dto.Product {

	filt := expression.Name("Title").Contains(searchTerm)

	proj := expression.NamesList(
		expression.Name("ID"),
		expression.Name("Title"),
		expression.Name("Price"),
	)

	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	errors.HandleIfError(err)

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(config.GetString("dynamodb.productsTableName")),
	}

	result, err := svc.Client.Scan(params)
	errors.HandleIfError(err)

	products := []dto.Product{}
	for _, i := range result.Items {
		item := dto.Product{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		errors.HandleIfError(err)

		products = append(products, item)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].Price > products[j].Price
	})

	return products
}

// ProductSearch searches products for matches
func ProductSearch(search *database.Search, searchTerm string) []dto.Product {
	params := &cloudsearchdomain.SearchInput{
		Query: aws.String(searchTerm),
		Sort:  aws.String("price desc"),
	}
	result, err := search.Client.Search(params)
	errors.HandleIfError(err)

	products := []dto.Product{}
	for _, i := range result.Hits.Hit {
		price, err := strconv.ParseFloat(*i.Fields["price"][0], 64)
		errors.HandleIfError(err)
		item := dto.Product{
			ID:    *i.Id,
			Title: *i.Fields["title"][0],
			Price: price,
		}

		products = append(products, item)
	}
	return products
}
