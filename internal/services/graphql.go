package services

import (
	"encoding/json"
	"errors"
	"sort"
	"task/internal/api/graphql/graph/model"

	"net/http"

	"github.com/allegro/bigcache/v3"
)

type GraphQLService struct {
	Cache *bigcache.BigCache
}

const (
	apiMeterial = "https://jgpjqcuk9e.execute-api.us-east-2.amazonaws.com/materials"
	apiSupplier = "https://jgpjqcuk9e.execute-api.us-east-2.amazonaws.com/suppliers"
)

type Material struct {
	ID           int     `json:"id"`
	MaterialName string  `json:"materialName"`
	MaterialType string  `json:"materialType"`
	Price        float64 `json:"price"`
	Unit         string  `json:"unit"`
	Rating       int     `json:"rating"`
	Quality      int     `json:"quality"`
}

type Supplier struct {
	ID               int                      `json:"id"`
	SupplierName     string                   `json:"supplierName"`
	SupplierLocation string                   `json:"supplierLocation"`
	Materials        map[string][]StockDetail `json:"materials"`
}

type StockDetail struct {
	MaterialName      string `json:"materialName"`
	StockAvailability string `json:"stockAvailability"`
	StockQuantity     int    `json:"stockQuantity"`
}

type BestMaterial struct {
	Material Material `json:"material"`
	Supplier Supplier `json:"supplier"`
}

func fetchMaterials(cache *bigcache.BigCache) ([]Material, error) {
	cachedData, err := cache.Get("materials")
	if err == nil {
		// if cache hit, return cached data
		var cachedMaterials []Material
		if err := json.Unmarshal(cachedData, &cachedMaterials); err == nil {

			return cachedMaterials, nil
		}
	}

	// fetch from apies if cache miss
	var allMaterials []Material
	for i := 0; i < 5; i++ {
		resp, err := http.Get(apiMeterial)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var materials []Material
		if err := json.NewDecoder(resp.Body).Decode(&materials); err != nil {
			return nil, err
		}

		allMaterials = append(allMaterials, materials...)
	}

	// storre fetched data in cache
	dataToCache, _ := json.Marshal(allMaterials)
	cache.Set("materials", dataToCache)

	return allMaterials, nil
}

func fetchSuppliers(locality string) ([]Supplier, error) {
	resp, err := http.Get(apiSupplier)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var suppliers []Supplier
	err = json.NewDecoder(resp.Body).Decode(&suppliers)
	if err != nil {
		return nil, err
	}

	// =filter suppliers by locality
	var filteredSuppliers []Supplier
	for _, supplier := range suppliers {
		if supplier.SupplierLocation == locality {
			filteredSuppliers = append(filteredSuppliers, supplier)
		}
	}

	// If no suppliers found in the given locality, return an error
	if len(filteredSuppliers) == 0 {
		return nil, errors.New("no suppliers found in the specified locality")
	}

	return filteredSuppliers, nil
}

func (r *GraphQLService) FindBestMaterial(materialType string, price float64, locality string) ([]*model.BestMaterial, error) {
	materials, err := fetchMaterials(r.Cache)
	if err != nil {
		return nil, err
	}

	suppliers, err := fetchSuppliers(locality)
	if err != nil {
		return nil, err
	}

	// first price of meterial should be lesss than or equal to price enetered by user
	var filteredMaterials []Material
	for _, mat := range materials {
		if mat.MaterialType == materialType {
			filteredMaterials = append(filteredMaterials, mat)
		}
	}

	// arranging based on qualiity
	sort.SliceStable(filteredMaterials, func(i, j int) bool {
		if filteredMaterials[i].Quality == filteredMaterials[j].Quality {
			return filteredMaterials[i].Rating > filteredMaterials[j].Rating
		}
		return filteredMaterials[i].Quality > filteredMaterials[j].Quality
	})

	if len(filteredMaterials) == 0 {
		return nil, errors.New("no suitable material found")
	}

	return mergeMaterialsAndSuppliers(filteredMaterials, suppliers, materialType)

}

func mergeMaterialsAndSuppliers(materials []Material, suppliers []Supplier, materialType string) ([]*model.BestMaterial, error) {
	var mergedList []*model.BestMaterial

	for _, material := range materials {

		for _, supplier := range suppliers {

			for _, stock := range supplier.Materials[materialType] {
				if stock.MaterialName == material.MaterialName {
					mergedList = append(mergedList, &model.BestMaterial{
						Material: &model.Material{
							ID:           int32(material.ID),
							MaterialName: material.MaterialName,
							MaterialType: material.MaterialType,
							Price:        material.Price,
							Unit:         material.Unit,
							Rating:       int32(material.Rating),
							Quality:      int32(material.Quality),
						},
						Supplier: &model.Supplier{
							ID:               int32(supplier.ID),
							SupplierName:     supplier.SupplierName,
							SupplierLocation: supplier.SupplierLocation,
						},
					})
				}
			}
		}
	}

	// if no matches found, return an error
	if len(mergedList) == 0 {
		return nil, errors.New("no matching materials found with available suppliers")
	}

	return mergedList, nil
}
