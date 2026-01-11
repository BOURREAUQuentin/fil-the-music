db = db.getSiblingDB("catalog_db");
db.tracks.insertMany([
  {
    track_id: "trk_112233",
    title: "One More Time",
    artist_id: "art_778899",
    album_name: "Discovery",
    release_date: "2001-03-12",
  },
  {
    track_id: "trk_445566",
    title: "D.A.N.C.E",
    artist_id: "art_123456",
    album_name: "Cross",
    release_date: "2007-06-11",
  },
]);

db.artists.insertMany([
  {
    artist_id: "art_778899",
    name: "Daft Punk",
    genres: ["Electronic", "French House", "Synth-pop"],
  },
  {
    artist_id: "art_123456",
    name: "Justice",
    genres: ["Electro", "Indie Dance"],
  },
]);
