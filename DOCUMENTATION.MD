
Giobby API
Home
Api Explorer
Documentation
Reference
Login Api-client
Questa modalità è pensata per far fare il login su Giobby agli utenti di un'applicazione di terze parti, così da autorizzare quest'ultima a fare chiamate a Giobby per conto dell'utente stesso.

In questo caso le API di Giobby utilizzano OpenID Connect come protocollo d'autenticazione.

Primo step: indirizzare l'utente alla pagina di login di Giobby linkandolo sul seguente url

baseUrl = "https://auth.giobby.com/auth/realms/api-client/protocol/openid-connect/auth";
urlLogin = "{baseUrl}?client_id={client_id}&redirect_uri={redirect_uri}&response_mode=fragment&
response_type=code&scope=openid";

Una volta che l'utente si sarà autenticato con successo verrà rediretto alla pagina di redirect ({redirect_uri}) e nell'url verrà inserito il parametro code che servirà per il secondo step.

Secondo step: prelevare dall'url il parametro "code".
Effettuare una chiamata POST al seguente url

https://auth.giobby.com/auth/realms/api-client/protocol/openid-connect/token
passando nel body i seguenti parametri

"code" = {code},
"grant_type" = "authorization_code",
"client_id" = {client_id},
"redirect_uri" = {redirect_uri}
In caso di successo la risposta prevede un json di questo tipo:

{"access_token": ...,
"expires_in": ...,
"refresh_expires_in": ...,
"refresh_token": ...,
"token_type": ...,
"id_token": ...,
"not-before-policy": ...,
"session_state": ...}

Sotto questa modalità l'access_token ha una durata limitata ("expires_in" -> valore in secondi) e una volta scaduto bisogna eseguire il refresh del token.

Per farlo, effettuare una chiamata POST al seguente url

https://auth.giobby.com/auth/realms/api-client/protocol/openid-connect/token
passando i seguenti parametri nel body della richiesta:

"grant_type" = "refresh_token",
"refresh_token" =  {refresh_token},
"client_id" = {client_id}
In caso di successo la risposta prevede un json con struttura identica al precedente.

Per questa modalità consigliamo di utilizzare uno tra gli adapter di Keycloak per facilitare il processo di autenticazione. Tutto il necessario lo si può trovare seguendo il seguente link: Keycloak OpenID Connect documentation.

Cliccando sul link di seguito avvierai il download di un progetto di esempio che utilizza la libreria javascript di keycloak.
Download progetto esempio api-client


Login Api-server
Questa modalità è pensata per i clienti di Giobby proprietari di app di terze parti che vogliono utilizzare le API con il loro cid.


Primo step: effettuare una chiamata POST al seguente url

https://auth.giobby.com/auth/realms/api-server/protocol/openid-connect/token
passando nel body di richiesta i seguenti parametri:

{
    "grant_type" : "password",
    "client_id" : {client_id},
    "username" : {username},
    "cid" : {cid},
    "password" : {password}
}

In caso di successo la risposta prevede un json di questo tipo:

{"access_token": ...,
"expires_in": ...,
"refresh_expires_in": ...,
"refresh_token": ...,
"token_type": ...,
"not-before-policy": ...,
"session_state": ...}

Esempio di chiamata curl:

curl -d "client_id={client_id}" -d "cid={cid}"
-d "username={username}" -d "password={password}"
-d "grant_type=password"
"https://auth.giobby.com/auth/realms/api-server/protocol/openid-connect/token"
                                    
Cliccando sul link di seguito avvierai il download di un progetto di esempio.
Download progetto esempio api-server


Parte comune
Ad autenticazione completa, la restante parte è la stessa per entrambe le modalità.

Lo step successivo è quello di ottenere l'endpoint per effettuare le chiamate alle API.

Bisogna effettuare una chiamata GET all'url https://app.giobby.com/GiobbyApiLogin/v1/endpoint passando come queryparam idCompany={idCompany}. Per ottenere questo parametro si deve decodificare l'access_token dell'utente. Essendo in formato JWT, una volta decodificato, dal json risultante è possibile recuperare l'idCompany.
Esempio completo:

https://app.giobby.com/GiobbyApiLogin/v1/endpoint?idCompany=00000

In risposta alla suddetta chiamata è previsto un file json di questo tipo:

{"responseCode":200,
"errorCode":0,
"developerMessage":"",
"userMessage":"",
"concatWarningMessages":"",
"GiobbyApiURL": ...}
Il campo "GiobbyApiURL" rappresenta l'endpoint da chiamare per l'utilizzo delle API e deve essere recuperato ad ogni login di un utente.


Esempio chiamata API
Volendo quindi chiamare per esempio l'API che restituisce i dati dell'utente loggato, si andrà ad effettuare la seuguente chiamata:

curl -X GET --header 'Accept: application/json' --header 'Authorization: Bearer {access_token}' --header 'X-Giobby-Realm: {realm}' '{GiobbyApiURL}/loggeduser'
dove i vari header sono obbligatori.

ATTENZIONE: {realm} acquisisce il vaore di 'api-server' o 'api-client' a seconda della modalità utilizzata per l'autenticazione.