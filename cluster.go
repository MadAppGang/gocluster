// goluster is very fast library for  geospatial point clustering server side (or client side)
// The cluster use hierarchical greedy clustering approach.
// The same approach used by Dave Leaver with his fantastic Leaflet.markercluster plugin.
// So this approach is extremely fast, the only drawback is that all clustered points are stored in memory
// This library is deeply inspired by MapBox's superclaster JS library and blogpost: https://www.mapbox.com/blog/supercluster/
//
package cluster

// Cluster struct get a list or stream of geo objects
// and produce all levels of clusters
// MinZoom - minimum  zoom level to generate clusters
// MaxZoom - maximum zoom level to generate clusters
// Zoom range is limited by 0 to 21, and MinZoom could not be larger, then MaxZoom
// PointSize - pixel size of marker, affects clustering radius
// TileSize - size of tile in pixels, affects clustering radius
// MinBranch - Minimum Branching factor
// MaxBranch - Minimum Branching factor
// minimum and maximum branching factor are settings for index, you could play with it to tweak performance for your case
type Cluster struct {
	MinZoom   int
	MaxZoom   int
	PointSize int
	TileSize  int
	MinBranch int
	MaxBranch int
}

// Create new Cluster instance with default parameters:
// MinZoom = 0
// MaxZoom = 18
// PointSize = 50
// TileSize = 512 (GMaps and OSM default)
// MinBranch = 32
// MinBranch = 64
func NewCluster() Cluster {
	return Cluster{
		MinZoom:0,
		MaxZoom:18,
		PointSize:50,
		TileSize:512,
		MinBranch:32,
		MaxBranch:64,
	}
}




