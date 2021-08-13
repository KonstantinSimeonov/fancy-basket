readonly JSON='{ "Name": "Pencho", "Email": "pencho@penchomail.pencho", "Password": "zdr_proateli1" }' 
readonly JSON2='{ "Name": "Stoika", "Email": "stoika@gmail.com", "Password": "zdr_proateli2" }' 
curl -d "$JSON2" localhost:4000/users
token=$(curl -d "$JSON2" localhost:4000/tokens)
echo $token

curl -H "Authorization: $token" -d '{ "CategoryID": "1", "Name": "Shnorhel sys PANDELKA", "Price": 6.14 }' localhost:4000/products
