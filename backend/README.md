# Backend de Go

Este es el backend de la aplicación, desarrollado en Go utilizando el framework Gin. Sigue los principios de la Arquitectura Limpia para una mejor mantenibilidad y separación de responsabilidades.

## Requisitos previos

- [Go](https://golang.org/doc/install) (versión 1.24.3 o superior)
- [Docker](https://docs.docker.com/get-docker/) y [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)

## Configuración

1.  Crea un archivo de entorno para el desarrollo local. Puedes copiar el de ejemplo:
    ```bash
    cp dev.env.example dev.env
    ```
2.  Abre el archivo `dev.env` y ajusta las variables de entorno según sea necesario. Como mínimo, deberás configurar las credenciales de la base de datos PostgreSQL. También puedes configurar la variable `PORT` para cambiar el puerto en el que se ejecuta la aplicación (el valor predeterminado es `3000`).

## Cómo ejecutar la aplicación

Hay dos formas de ejecutar la aplicación: localmente para el desarrollo o a través de Docker.

### Ejecución local (con live reload)

Este es el método recomendado para el desarrollo, ya que utiliza `air` para recargar la aplicación automáticamente cuando se detectan cambios en el código.

1.  Asegúrate de que la base de datos PostgreSQL esté en funcionamiento. Puedes iniciarla con Docker Compose:
    ```bash
    docker-compose up -d db
    ```
2.  Ejecuta la aplicación:
    ```bash
    make watch
    ```

La aplicación estará disponible en `http://localhost:3000`.

### Ejecución con Docker

Este método ejecuta la aplicación en un contenedor de Docker, lo que es ideal para un entorno de producción o para simularlo.

1.  Construye y levanta los contenedores (la aplicación y la base de datos):
    ```bash
    docker-compose up --build
    ```

La aplicación estará disponible en `http://localhost:3000`.

## Endpoints de la API

La API está versionada bajo el prefijo `/api/v1`.

- `GET /health`: Endpoint de comprobación de estado. Devuelve un `200 OK` si el servidor está en funcionamiento.
- `POST /messages`: Crea un nuevo mensaje. El cuerpo de la petición debe ser un JSON con el siguiente formato:
  ```json
  {
    "text": "Hola, mundo!",
    "authorId": "un-id-de-autor",
    "channelId": "un-id-de-canal"
  }
  ```

## Estructura del proyecto

El proyecto sigue la Arquitectura Limpia y está organizado en las siguientes capas:

- `domain`: Contiene las entidades y la lógica de negocio principal.
- `application`: Contiene los casos de uso de la aplicación y las interfaces de los repositorios.
- `infrastructure`: Contiene las implementaciones de las interfaces, como la conexión a la base de datos y los repositorios.
- `presentation`: Contiene los controladores de la API y la configuración del servidor HTTP (Gin).