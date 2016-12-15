// goluster is very fast library for  geospatial point clustering server side (or client side)
// The cluster use hierarchical greedy clustering approach.
// The same approach used by Dave Leaver with his fantastic Leaflet.markercluster plugin.
// So this approach is extremely fast, the only drawback is that all clustered points are stored in memory
// This library is deeply inspired by MapBox's superclaster JS library and blog post: https://www.mapbox.com/blog/supercluster/
//
package cluster

import (
	"math"
	"github.com/dhconnelly/rtreego"
)


//That Zoom level indicate impossible large zoom level (Cluster's max is 21)
const InfinityZoomLevel = 100


// GeoCoordinates represent position in the Earth
type GeoCoordinates struct {
	Lon float64
	Lat float64
}

// all object, that you want to cluster should implement this protocol
type GeoPoint interface {
	GetCoordinates() GeoCoordinates
	// Any kind of identification, will be placed in result index instead of object
	// If you do not want to store the whole object
	GeoPointID() string
}

// Interface of object that Cluster use to create new clustered point
// You could implement you own class and create your own objects, or use SimpleClusteredPointsProducer
type ClusteredPointsProducer interface {
	// Construct clustered point from geo point
	NewPoint(geoPoint GeoPoint) ClusteredPoint
	NewPointWithPoints(points []ClusteredPoint, x,y float64, zoom int) ClusteredPoint
}


// The objects, that you are getting as the result of clustering
// They could store GeoPoint or just id of geo point, depends oin your own implementation
type ClusteredPoint interface {
	//Get and set mercator projection coordinates and zoom level
	GetXYZ() (float64, float64, int)
	SetXY(x, y float64)
	SetZoom(zoom int)

	//Number of points in this cluster, 1 or more
	PointsCount() int

	//Implements spatial interface of rtreego
	Bounds() *rtreego.Rect
}



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
// DeepLink is flag, if it's true, all resulting ClusteredPoints will have original points in it
// if DeepLink is false, just id is stored, saving you memory and some time
type Cluster struct {
	MinZoom   int
	MaxZoom   int
	PointSize int
	TileSize  int
	MinBranch int
	MaxBranch int
	Indexes []*rtreego.Rtree
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
		MinZoom:   0,
		MaxZoom:   18,
		PointSize: 50,
		TileSize:  512,
		MinBranch: 32,
		MaxBranch: 64,
	}
}

// ClusterPoint get points and create multilevel clustered indexes
// this method use SimpleClusteredPointsProducer to produce clustered points
// so you will get SimpleClusteredPoints as result
// and DeepLink will be false
func (c *Cluster) ClusterPoints(points []GeoPoint) error {
	return c.ClusterPointsWithDeepLinking(points, false)
}

// ClusterPointsWithDeepLinking get points and create multilevel clustered indexes
// this method use SimpleClusteredPointsProducer to produce clustered points
// so you will get SimpleClusteredPoints as result
// you could set DeepLink here as second param for  SimpleClusteredPointsProducer
func (c *Cluster) ClusterPointsWithDeepLinking(points []GeoPoint, deepLink bool) error {
	pp := SimpleClusteredPointsProducer{}
	pp.DeepLink = deepLink
	return c.ClusterPointsWithProducer(points, &pp)
}

// ClusterPointsWithProducer get points and create multilevel clustered indexes
// you have to provide initialised point producer here
func (c *Cluster) ClusterPointsWithProducer(points []GeoPoint, pointProducer ClusteredPointsProducer) error {

	//limit max Zoom
	if c.MaxZoom > 21 { c.MaxZoom = 21 }
	//adding extra layer for infinite zoom (initial) layers data storage
	c.Indexes = make([]*rtreego.Rtree, c.MaxZoom-c.MinZoom+2)

	clusters := translatePointsToClusters(points, pointProducer)
	for z := c.MaxZoom; z >= c.MinZoom; z-- {

		//create index from clusters from previous iteration
		idx := rtreego.NewTree(2, c.MinBranch, c.MaxBranch)
		for _, p := range clusters {
			idx.Insert(p)
		}
		c.Indexes[z+1] = idx
		//create clusters for level up using just created index
		clusters = c.clusterize(clusters, z, pointProducer)
	}

	//index topmost points
	idx := rtreego.NewTree(2, c.MinBranch, c.MaxBranch)
	for _, p := range clusters {
		idx.Insert(p)
	}
	c.Indexes[c.MinZoom] = idx
	return nil
}

func (c *Cluster)GetTile(x,y,z int) []ClusteredPoint {

}


//clusterize points for zoom level
func (c *Cluster)clusterize(clusters []ClusteredPoint, zoom int, pp ClusteredPointsProducer) []ClusteredPoint {
	var result []ClusteredPoint
	var r float64 = float64(c.PointSize) / float64( c.TileSize * (1 << uint(zoom)))

	//iterate all clusters
	for _, p :=  range clusters {
		//skip points we have already clustered
		if _, _, z := p.GetXYZ(); z <= zoom {
			continue
		} else {
			p.SetZoom(z)
		}

		//find all neighbours
		tree := c.Indexes[zoom+1]

		wx, wy, _ := p.GetXYZ()
		bbox := rtreego.Point{ wx, wy }.ToRect(r / 2.0)
		neighbours := tree.SearchIntersect(bbox)

		found := false
		nPoints := p.PointsCount()
		wx = wx * float64(nPoints)
		wy = wy * float64(nPoints)

		var foundNeighbours []ClusteredPoint
		for _, nb := range neighbours {
			b := nb.(ClusteredPoint)
			bx, by, bz := b.GetXYZ()
			if ((zoom) < bz) && (dist(p,b) < r) {
				found = true
				wx += bx * float64(b.PointsCount())
				wy += by * float64(b.PointsCount())
				nPoints += nPoints
				b.SetZoom(zoom) //set the zoom to skip in other iterations
				foundNeighbours = append(foundNeighbours, b)
			}
		}

		var newClaster ClusteredPoint
		if found {
			newClaster = p
		} else {
			wx = wx / float64(nPoints)
			wy = wy / float64(nPoints)
			newClaster = pp.NewPointWithPoints(foundNeighbours, wx, wy, InfinityZoomLevel)
		}
		result = append(result, newClaster)
	}
	return result
}

////////// End of Cluster implementation


// Simple ClusteredPointsProducer implementation, produce SimpleClusteredPoint
// if you set DeepLink to true, it will include GeoPoint in SimpleClusteredPoint
// if DeepLink is false, store only array of identificators
type SimpleClusteredPointsProducer struct {
	// flag that indicate, that GeoPoint should be included in produced Clustered points
	DeepLink  bool
}

//implementation of protocols requirement, produce SimpleClusteredPoint from GeoPoint
func (pp *SimpleClusteredPointsProducer)NewPoint(geoPoint GeoPoint) ClusteredPoint {
	cp := SimpleClusteredPoint{}
	if pp.DeepLink == true {
		cp.GeoPointObjects = []GeoPoint{geoPoint}
	}
	cp.GeoPointIDs = []string{geoPoint.GeoPointID()}
	cp.Coordinates = geoPoint.GetCoordinates()
	cp.X, cp.Y = mercatorProjection(cp.Coordinates)
	cp.Zoom = InfinityZoomLevel
	return &cp
}

func (pp *SimpleClusteredPointsProducer) NewPointWithPoints(points []ClusteredPoint, x,y float64, zoom int) ClusteredPoint {
	cp := SimpleClusteredPoint{}
	for _, gp := range points {
		p := gp.(*SimpleClusteredPoint)
		p.SetXY(x,y)
		p.SetZoom(zoom)
		if pp.DeepLink {
			cp.GeoPointObjects = append(cp.GeoPointObjects, p.GeoPointObjects...)
		}
		cp.GeoPointIDs = append( cp.GeoPointIDs, cp.GeoPointIDs...)
	}
	return &cp
}

// Basic implementation of clustered point,
// you could compose this struct in your own type
// SimpleClusteredPoint embedding this type
type BasicClusteredPoint struct {
	//Mercator projection to map
	X float64
	//Mercator projection to map
	Y float64
	Zoom        int
}


func (bcp *BasicClusteredPoint)GetXYZ() (float64, float64, int) {
	return bcp.X, bcp.Y, bcp.Zoom
}

func (bcp *BasicClusteredPoint)SetXY(x, y float64) {
	bcp.X = x
	bcp.Y = y
}

func (bcp *BasicClusteredPoint)SetZoom(zoom int) {
	bcp.Zoom = zoom
}

func (bcp *BasicClusteredPoint)PointsCount() int{
	return 0
}


func (bcp *BasicClusteredPoint)Bounds() *rtreego.Rect{
	const df = 0.000001 //11cm
	return rtreego.Point{bcp.X, bcp.Y}.ToRect(df)
}




// General and simple implementation of  ClusteredPoint interface
// SimpleClusteredPoint is produced by SimpleClusteredPointsProducer
// But you could use it as well
type SimpleClusteredPoint struct {
	BasicClusteredPoint
	//Coordinates
	Coordinates GeoCoordinates
	GeoPointIDs     []string
	GeoPointObjects []GeoPoint
}

//override basic implementation, that always returns 0
func (cp *SimpleClusteredPoint)PointsCount() int{
	return len(cp.GeoPointIDs)
}


//////
// private stuff
func translatePointsToClusters(points []GeoPoint, pointProducer ClusteredPointsProducer) []ClusteredPoint {
	var result = make([]ClusteredPoint, len(points))
	for i, p := range points {
		cp := pointProducer.NewPoint(p)
		cp.SetZoom(InfinityZoomLevel)
		result[i] = cp
	}
	return result
}


func mercatorProjection(coordinates GeoCoordinates) (float64, float64) {
	const mercatorPole = 20037508.34
	x := mercatorPole / 180.0 * coordinates.Lon
	y := math.Log(math.Tan((90.0 + coordinates.Lat) * math.Pi / 360.0)) / math.Pi * mercatorPole
	y = math.Max(-mercatorPole, math.Min(y, mercatorPole))
	return x, y
}

// Dist computes the Euclidean distance between two points p and q.
func dist(q, p ClusteredPoint) float64 {
	x1, y1, _ := p.GetXYZ()
	x2, y2, _ := q.GetXYZ()
	dx := x1-x2
	dy := y1-y2
	sum := (dx*dx)+(dy*dy)
	return math.Sqrt(sum)
}
