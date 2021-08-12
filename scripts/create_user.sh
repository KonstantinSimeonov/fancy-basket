readonly JSON='{ "Name": "Pencho", "Email": "pencho@penchomail.pencho", "Password": "zdr_proateli1" }' 
# curl -d "$JSON" localhost:4000/users
token=$(curl -d "$JSON" localhost:4000/tokens)
echo $token

curl -H "Authorization: $token" -d '{ "CategoryID": "1", "Name": "Shnorhel sys zvynche", "Price": 3.14 }' localhost:4000/products
