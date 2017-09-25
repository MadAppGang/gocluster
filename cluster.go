package cluster

import (
	"math"

	"github.com/MadAppGang/kdbush"
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
}

type ClusterCustomizer interface {
	GeoPoint2ClusterPoint(point GeoPoint) ClusterPoint
	AggregateClusterPoints(point ClusterPoint, aggregated []ClusterPoint, zoom int) ClusterPoint
}

type defaultCustomizer struct {

}

func (dc defaultCustomizer) GeoPoint2ClusterPoint(point GeoPoint) ClusterPoint {
	cp := clusterPoint{}
	return cp
}

func (dc defaultCustomizer) AggregateClusterPoints(point ClusterPoint, aggregated []ClusterPoint, zoom int) ClusterPoint {
	return point
}


//Struct that implements clustered points
//could have only one point or set of points
type ClusterPoint interface {
	Coordinates() (float64, float64)
	getX() float64
	getY() float64
	setX(float64) ClusterPoint
	setY(float64) ClusterPoint
	setZoom(int) ClusterPoint
	getZoom() int
	setNumPoints(int) ClusterPoint
	getNumPoints() int
	setId(int) ClusterPoint
}

type clusterPoint struct {
	X,Y float64
	zoom int
	Id int //Index for pint, Id for cluster
	NumPoints int
}


func (cp clusterPoint)	Coordinates() (float64, float64) {
	return cp.X, cp.Y
}

func (cp clusterPoint) getX() float64 {
	return cp.X
}

func (cp clusterPoint) getY() float64 {
	return cp.Y
}

func (cp clusterPoint) setX(x float64) ClusterPoint {
	cp.X = x
	return cp
}

func (cp clusterPoint) setY(y float64) ClusterPoint {
	cp.Y = y
	return cp
}

func (cp clusterPoint) setZoom(zoom int) ClusterPoint {
	cp.zoom = zoom
	return cp
}
func (cp clusterPoint) getZoom() int{
	return cp.zoom
}

func (cp clusterPoint) setNumPoints(numPoints int) ClusterPoint {
	cp.NumPoints = numPoints
	return cp
}
func (cp clusterPoint) getNumPoints() int{
	return cp.NumPoints
}
func (cp clusterPoint) setId(id int) ClusterPoint {
	cp.Id = id
	return cp
}







// Cluster struct get a list or stream of geo objects
// and produce all levels of clusters
// MinZoom - minimum  zoom level to generate clusters
// MaxZoom - maximum zoom level to generate clusters
// Zoom range is limited by 0 to 21, and MinZoom could not be larger, then MaxZoom
// PointSize - pixel size of marker, affects clustering radius
// TileSize - size of tile in pixels, affects clustering radius
type Cluster struct {
	MinZoom   int
	MaxZoom   int
	PointSize int
	TileSize  int
	NodeSize  int
	Indexes   []*kdbush.KDBush
	Points    []GeoPoint
	ClusterCustomizer ClusterCustomizer
	ClusterIdxSeed int
	clusterIDLast int
}


// Create new Cluster instance with default parameters:
// MinZoom = 0
// MaxZoom = 16
// PointSize = 40
// TileSize = 512 (GMaps and OSM default)
// NodeSize is size of the KD-tree node, 64 by default. Higher means faster indexing but slower search, and vise versa.
func NewCluster() *Cluster {
	return &Cluster{
		MinZoom:   0,
		MaxZoom:   16,
		PointSize: 40,
		TileSize:  512,
		NodeSize:  64,
		ClusterCustomizer: defaultCustomizer{},
	}
}

func NewClusterFromCustomizer (customizer ClusterCustomizer) *Cluster {
	return &Cluster{
		MinZoom:   0,
		MaxZoom:   16,
		PointSize: 40,
		TileSize:  512,
		NodeSize:  64,
		ClusterCustomizer: customizer,
	}
}


// ClusterPoint get points and create multilevel clustered indexes
// All points should implement GeoPoint interface
// they are not copied, so you could not worry about memory efficiency
// And GetCoordinates called only once for each object, so you could calc it on the fly, if you need
func (c *Cluster) ClusterPoints(points []GeoPoint) error {

	//limit max Zoom
	if c.MaxZoom > 21 { c.MaxZoom = 21 }

	//adding extra layer for infinite zoom (initial) layers data storage
	c.Indexes = make([]*kdbush.KDBush, c.MaxZoom-c.MinZoom+2)
	c.Points = points

	//get digits number, start from next exponent
	//if we have 78, all cluster will start from 100...
	//if we have 986 points, all clusters ids will start from 1000
	c.ClusterIdxSeed = int(math.Pow(10,float64(digitsCount(len(points)))))
	c.clusterIDLast = c.ClusterIdxSeed


	clusters := c.translateGeoPointsToClusterPoints(points)

	for z := c.MaxZoom; z >= c.MinZoom; z-- {

		//create index from clusters from previous iteration
		c.Indexes[z+1] = kdbush.NewBush(clustersToPoints(clusters), c.NodeSize)

		//create clusters for level up using just created index
		clusters = c.clusterize(clusters, z)
	}

	//index topmost points
	c.Indexes[c.MinZoom] = kdbush.NewBush(clustersToPoints(clusters), c.NodeSize)
	return nil
}

// Returns the array of clusters for zoom level.
// The northWest and southEast points are boundary points of square, that should be returned.
// nothWest is left topmost point.
// southEast is right bottom point.
// return the object for clustered points,
// X coordinate of returned object is Longitude and
// Y coordinate of returned object is Latitude
func (c *Cluster)GetClusters(northWest, southEast GeoPoint, zoom int) []ClusterPoint {
	index := c.Indexes[c.limitZoom(zoom)]
	nwX, nwY := MercatorProjection(northWest.GetCoordinates())
	seX, seY := MercatorProjection(southEast.GetCoordinates())
	ids := index.Range(seX, seY,nwX,nwY)
	var result []ClusterPoint = make ([]ClusterPoint, len(ids))
	for i := range ids {
		p := index.Points[ids[i]].(ClusterPoint)
		coordinates :=  ReverseMercatorProjection(p.getX(), p.getY())
		p = p.setX(coordinates.Lon)
		p = p.setY(coordinates.Lat)
		result[i] = p
	}

	return result
}

// AllClusters returns all cluster points, array of ClusterPoint,  for zoom on the map.
// X coordinate of returned object is Longitude and.
// Y coordinate of returned object is Latitude.
func (c *Cluster)AllClusters(zoom int) []ClusterPoint {
	index := c.Indexes[c.limitZoom(zoom)]
	points := index.Points
	var result []ClusterPoint = make ([]ClusterPoint, len(points))
	for i := range points {
		p := index.Points[i].(ClusterPoint)
		coordinates :=  ReverseMercatorProjection(p.getX(), p.getY())
		p = p.setX(coordinates.Lon)
		p = p.setY(coordinates.Lat)
		result[i] = p
	}

	return result
}


//return points for  Tile with coordinates x and y and for zoom z
// return objects with pixel coordinates
func (c *Cluster)GetTile(x,y,z int) []ClusterPoint {
	return c.getTile(x,y,z,false)
}

//return points for  Tile with coordinates x and y and for zoom z
// return objects with LatLon coordinates
func (c *Cluster)GetTileWithLatLon(x,y,z int) []ClusterPoint {
	return c.getTile(x,y,z,true)
}

func (c *Cluster)getTile(x,y,z int, latlon bool) []ClusterPoint {
	index := c.Indexes[c.limitZoom(z)]
	z2 := 1 << uint(z)
	z2f := float64(z2)
	extent := c.TileSize
	r := c.PointSize
	p := float64(r) / float64(extent)
	top := (float64(y) - p)/z2f
	bottom := (float64(y)+1+p) / z2f

	resultIds := index.Range(
		(float64(x)-p)/ z2f ,
		float64(top),
		(float64(x)+1+p)/z2f,
		bottom,
	)
	var result []ClusterPoint
	if latlon == true {
		result = c.pointIDToLatLonPoint(resultIds,index.Points)
	} else {
		result = c.pointIDToMerkatorPoint(resultIds,index.Points,float64(x),float64(y),z2f)
	}

	if (x == 0) {
		minX1 := float64(1-p)/z2f
		minY1 := float64(top)
		maxX1 := 1.0
		maxY1 := float64(bottom)
		resultIds = index.Range(minX1, minY1, maxX1, maxY1)
		var sr1 []ClusterPoint
		if latlon == true {
			sr1 = c.pointIDToLatLonPoint(resultIds, index.Points)
		} else {
			sr1 = c.pointIDToMerkatorPoint(resultIds, index.Points, z2f, float64(y), z2f)
		}
		result =  append(result, sr1...)

	}

	if x == (z2-1) {
		minX2 := 0.0
		minY2 := float64(top)
		maxX2 := float64(p)/z2f
		maxY2 := float64(bottom)
		resultIds = index.Range(minX2, minY2, maxX2, maxY2)
		var sr2 []ClusterPoint
		if latlon == true {
			sr2 = c.pointIDToLatLonPoint(resultIds, index.Points)
		} else {
			sr2 = c.pointIDToMerkatorPoint(resultIds,index.Points,-1,float64(y),z2f)
		}
		result =  append(result, sr2...)
	}


	return result
}

//calc Point mercator projection regarding tile
func(c *Cluster) pointIDToMerkatorPoint(ids []int, points []kdbush.Point, x, y, z2 float64) []ClusterPoint {
	var result []ClusterPoint
	for i := range ids {
		p := points[ids[i]].(ClusterPoint)
		//translate our coordinate system to mercator
		p = p.setX(float64(round(float64(c.TileSize) *(p.getX()*z2-x))))
		p = p.setY(float64(round(float64(c.TileSize) *(p.getY()*z2-y))))
		p = p.setZoom(0)
		result = append(result,p)
	}
	return result
}
func(c *Cluster) pointIDToLatLonPoint(ids []int, points []kdbush.Point) []ClusterPoint {
	var result []ClusterPoint = make ([]ClusterPoint, len(ids))
	for i := range ids {
		p := points[ids[i]].(ClusterPoint)
		coordinates :=  ReverseMercatorProjection(p.getX(), p.getY())
		p = p.setX(coordinates.Lon)
		p = p.setY(coordinates.Lat)
		result[i] = p
	}
	return result
}



//clusterize points for zoom level
func (c *Cluster)clusterize(points []ClusterPoint, zoom int) []ClusterPoint {
	var result []ClusterPoint
	var r float64 = float64(c.PointSize) / float64( c.TileSize * (1 << uint(zoom)))

	//iterate all clusters
	for pi :=  range points {
		//skip points we have already clustered
		p := points[pi]
		if p.getZoom() <= zoom {
			continue
		}

		//mark this point as visited
		p = p.setZoom(zoom)

		points[pi] = p


		//find all neighbours
		tree := c.Indexes[zoom+1]
		neighbourIds := tree.Within(&kdbush.SimplePoint{X:p.getX(),Y:p.getY()},r)

		nPoints := p.getNumPoints()
		wx := p.getX() * float64(nPoints)
		wy := p.getY() * float64(nPoints)

		var foundNeighbours []ClusterPoint

		for j := range neighbourIds {
			b := points[neighbourIds[j]]

			//Filter out neighbours, that are already processed (and processed point "p" as well)
			if zoom < b.getZoom() {
				wx += b.getX() * float64(b.getNumPoints())
				wy += b.getY() * float64(b.getNumPoints())
				nPoints += b.getNumPoints()
				b = b.setZoom(zoom) //set the zoom to skip in other iterations
				foundNeighbours = append(foundNeighbours, b)
				points[neighbourIds[j]] = b
			}
		}
		newCluster := p

		//create new cluster
		if len(foundNeighbours)>0 {
			newCluster = c.ClusterCustomizer.AggregateClusterPoints(newCluster, foundNeighbours, zoom)
			newCluster = newCluster.setX(wx / float64(nPoints))
			newCluster = newCluster.setY(wy / float64(nPoints))
			newCluster = newCluster.setNumPoints(nPoints)
			newCluster = newCluster.setZoom(InfinityZoomLevel)
			newCluster = newCluster.setId(c.clusterIDLast)
			c.clusterIDLast += 1
		}
		result = append(result, newCluster)
	}
	return result
}

func (c *Cluster)limitZoom(zoom int) int {
	if zoom > c.MaxZoom+1 { zoom = c.MaxZoom+1 }
	if zoom < c.MinZoom   { zoom = c.MinZoom }
	return zoom
}

////////// End of Cluster implementation

/////////////////////////////////
// private stuff
/////////////////////////////////

//translate geopoints to ClusterPoints witrh projection coordinates
func (c *Cluster) translateGeoPointsToClusterPoints(points []GeoPoint) []ClusterPoint {
	var result = make([]ClusterPoint, len(points))
	for i, p := range points {
		cp := c.ClusterCustomizer.GeoPoint2ClusterPoint(p)
		cp = cp.setZoom(InfinityZoomLevel)
		X, Y := MercatorProjection(p.GetCoordinates())
		cp = cp.setX(X)
		cp = cp.setY(Y)

		cp = cp.setNumPoints(1)

		cp = cp.setId(i)

		result[i] = cp
	}
	return result
}


// longitude/latitude to spherical mercator in [0..1] range
func MercatorProjection(coordinates GeoCoordinates) (float64, float64) {
	x := coordinates.Lon / 360.0 + 0.5
	sin := math.Sin(coordinates.Lat * math.Pi / 180.0)
	y := (0.5 - 0.25 * math.Log((1+sin) / (1-sin)) / math.Pi )
	if y < 0 { y = 0 }
	if y > 1 { y = 1 }
	return x, y
}
func ReverseMercatorProjection(x, y float64) GeoCoordinates {
	result := GeoCoordinates{}
	result.Lon = (x-0.5)*360
	y2 := (180 - y * 360) * math.Pi / 180.0
	result.Lat = 360 * math.Atan(math.Exp(y2))/math.Pi - 90
	return result
}


//count number of digits, for example 123356 will return 6
func digitsCount(a int) int {
	return  int(math.Floor(math.Log10( math.Abs (float64(a))))) + 1
}

func clustersToPoints(points []ClusterPoint) []kdbush.Point {
	result := make([]kdbush.Point, len(points))
	for i, v := range points { result[i] = v }
	return result
}

func round(val float64) int {
	if val < 0 {
		return int(val-0.5)
	}
	return int(val+0.5)
}