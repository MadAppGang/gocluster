// MIT License
//
// Copyright (c) 2016 MadAppGang

// A very fast Golang library for geospatial point clustering.
//
//
// gocluster is very fast library for  geospatial point clustering server side (or client side)
//
// The cluster use hierarchical greedy clustering approach.
//
// The same approach used by Dave Leaver with his fantastic Leaflet.markercluster plugin.
//
// So this approach is extremely fast, the only drawback is that all clustered points are stored in memory
//
// This library is deeply inspired by MapBox's superclaster JS library and blog post: https://www.mapbox.com/blog/supercluster/
//
// Very easy to use:
//	//1.Create new cluster
// 	c := NewCluster()
//
//	//2.Convert slice of your objects to slice of GeoPoint (interface) objects
//  	geoPoints := make([]GeoPoint, len(points))
//  	for i := range points {
//  		geoPoints[i] = points[i]
// 	}
//
//	//3.Build index
// 	c.ClusterPoints(geoPoints)
//
//	//3.Get tour tile with mercator coordinate projections to display directly on the map
// 	result := c.GetTile(0,0,0)
//
//
//
// Library has only one dependency, it's KD-tree geospatial index https://github.com/MadAppGang/kdbush
//
// All ids of ClusterPoint that you have as result are the index of initial array of Geopoint,
// so yu could get you point by this index
//
// Clusters of points are have autoincrement generated ids, started at ClusterIdxSeed
// ClusterIdxSeed is the next power of length of input array
//
// For example, if input slice of points length is 78,  ClusterIdxSeed == 100,
// if input slice of points length is 991,  ClusterIdxSeed == 1000
// etc
//
// TODO: Benchmarks
//
// TODO: demo server
package cluster
