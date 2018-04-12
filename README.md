# Gocluster

Please look at [godocs here](https://godoc.org/github.com/MadAppGang/gocluster).

Gocluster is a very fast Golang library for geospatial point clustering.

This image is demo of JS library, this will work faster, because Golang is faster :-)

TODO: implement goserver example

![clusters2](https://cloud.githubusercontent.com/assets/25395/11857351/43407b46-a40c-11e5-8662-e99ab1cd2cb7.gif)

The cluster use hierarchical greedy clustering approach.
The same approach used by Dave Leaver with his fantastic Leaflet.markercluster plugin.

So this approach is extremely fast, the only drawback is that all clustered points are stored in memory

This library is deeply inspired by MapBox's superclaster JS library and blog post: https://www.mapbox.com/blog/supercluster/

Very easy to use:

```go
//1.Create new cluster
c := NewCluster()
//2.Convert slice of your objects to slice of GeoPoint (interface) objects
geoPoints := make([]GeoPoint, len(points))
for i := range points {
  geoPoints[i] = points[i]
}
//3.Build index
c.ClusterPoints(geoPoints)
//4.Get tour tile with mercator coordinate projections to display directly on the map
result := c.GetTile(0,0,0)
```

Library has only one dependency, [it's KD-tree geospatial index](https://github.com/MadAppGang/kdbush)

All ids of `ClusterPoint` that you have as result are the index of initial array of Geopoint,
so you could get you point by this index.

Clusters of points are have autoincrement generated ids, started at `ClusterIdxSeed`.

`ClusterIdxSeed` is the next power of length of input array.
For example, if input slice of points length is `78`,  `ClusterIdxSeed == 100`,
if input slice of points length is `991`,  `ClusterIdxSeed == 1000`
etc

## Init cluster index

To init index, you need to prepare your data. All your points should implement `GeoPoint` interface:
```go
type GeoPoint interface {
	GetCoordinates() GeoCoordinates
}

type GeoCoordinates struct {
	Lon float64
	Lat float64
}
```

You could tweak the `Cluster`:

|parameter | default value | description |
|---|---|---|
|MinZoom | 0 | Minimum zoom level at which clusters are generated |
|MaxZoom | 16 | Minimum zoom level at which clusters are generated |
|PointSize | 40 | Cluster radius, in pixels |
|TileSize | 512 | Tile extent. Radius is calculated relative to this value |
|NodeSize | 64 | NodeSize is size of the KD-tree node. Higher means faster indexing but slower search, and vise versa. |

## Search point in boundary box

To search all  points inside the box, that are limited by the box, formed by north-west point and east-south points. You need to provide Z index as well.

```go

northWest := simplePoint{71.36718750000001, -83.79204408779539}
southEast := simplePoint{-71.01562500000001, 83.7539108491127}
zoom := 2
var result []ClusterPoint = c.GetClusters(northWest, southEast, zoom)

```

Returns the array of 'ClusterPoint' for zoom level.
Each point has following coordinates:
 * X coordinate of returned object is Longitude and
 * Y coordinate of returned object is Latitude
 * if the object is cluster of points (NumPoints > 1), the ID is generated started from ClusterIdxSeed (ID>ClusterIdxSeed)
 * if the object represents only one point, it's id is the index of initial GeoPoints array



## Search points for tile

OSM and Google maps [uses tiles system](https://developers.google.com/maps/documentation/javascript/maptypes#TileCoordinates) to optimize map loading.
So you could get all points for the tile with tileX, tileY and zoom:

```go
c := NewCluster()
c.ClusterPoints(geoPoints)
tileX := 0
tileY := 1
zoom := 4
result := c.GetTile(tileX, tileY, zoom)

```
In this case all coordinates are returned in pixels for that tile.
If you want to return objects with Lat, Long, use `GetTileWithLatLon` method.



TODO: Benchmarks
TODO: demo server
