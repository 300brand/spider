var keys =[];

var words = ['cloud'];

var pubIDs = [
	ObjectId('5272624d800b8e3d940006ed'),
	ObjectId('52726259800b8e3d94000bf9'),
	ObjectId('5294f834a75ae411660059b4'),
	ObjectId('52726255800b8e3d94000a4f'),
	ObjectId('5272625684e7533dd00002e3'),
	ObjectId('52726259800b8e3d94000bbb'),
	ObjectId('5272625aa75ae43e05000191'),
	ObjectId('5272625ba75ae43e050001a7'),
	ObjectId('5272625884e7533dd0000323'),
	ObjectId('52c74079800b8e3387005eb6'),
	ObjectId('52726256800b8e3d94000a8f'),
	ObjectId('5294d213800b8e10670010a9'),
	ObjectId('52a704d5800b8e67a0014867'),
	ObjectId('52a7045a84e753686f0151d4'),
	ObjectId('52726256800b8e3d94000aab'),
	ObjectId('52726256800b8e3d94000a87'),
	ObjectId('5272625ba75ae43e0500019d'),
	ObjectId('52726255800b8e3d94000a39'),
	ObjectId('5272625584e7533dd00002bf'),
	ObjectId('52726256800b8e3d94000a97'),
	ObjectId('52726256800b8e3d94000a67'),
	ObjectId('52726257800b8e3d94000af3'),
	ObjectId('52726254800b8e3d94000a05'),
	ObjectId('5272625884e7533dd0000327'),
	ObjectId('529501eeb6bbac2346003694'),
	ObjectId('5294faff84e75310ef00335e'),
	ObjectId('5272624684e7533dd00000cf'),
	ObjectId('52726256b6bbac3d58000147'),
	ObjectId('527262574113de3d8a0000fb'),
	ObjectId('527262494113de3d8a000051'),
	ObjectId('52c73eec84e7533371008784')
];

var _ = words.map(function(keyword){ return [ 20131115, 20140215 ].map(function(date){ keys.push({date: date, keyword: keyword}) }) })

var articleIDs = db.Keywords.aggregate(
  {
    $match: {
      _id: {
        $in: keys
      }
    }
  },
 
  {
    $project: {
      "keyword": "$_id.keyword",
      "article": "$articles"
    }
  },
 
  {
    $unwind: "$article"
  },
 
  {
    $group: {
      _id: "$article",
      phrase: { $push: "$keyword" },
      matched: { $sum: 1 }
    }
  },
 
  {
    $match: {
      matched: 1
    }
  }
).result.map(function(a) { return a._id; } );

db.getSiblingDB("300brand_Articles").Articles.find({ _id : { $in : articleIDs }, publicationid : { $in : pubIDs } });
