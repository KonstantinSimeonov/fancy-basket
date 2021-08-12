readonly JSON='{ "Name": "Pencho", "Email": "pencho@penchomail.pencho", "Password": "zdr_proateli1" }' 
# curl -d "$JSON" localhost:4000/users
curl -d "$JSON" localhost:4000/tokens
