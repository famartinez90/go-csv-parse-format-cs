# go-csv-parse-format-cs

## Objetivo

El objetivo es este programa es parsear las 71 votaciones de la cámara de diputados
y poder generar, a partir de las mismas, transacciones para luego, usando R y el 
algoritmo Apriori, obtener reglas interesantes.

## Compilación

```
cd parse/
go build
```

## Ejecución

```
cd parse/
./parse
```

Luego, se genera el archivo **transacciones.csv** con todas las transacciones.
