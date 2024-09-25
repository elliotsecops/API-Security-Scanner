**API Security Scanner**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Introducción

El API Security Scanner es una herramienta poderosa diseñada para ayudar a desarrolladores y profesionales de seguridad a evaluar la postura de seguridad de sus APIs. Realiza una serie de pruebas de seguridad, incluyendo verificaciones de autenticación, validación de métodos HTTP y detección de inyección SQL, para identificar posibles vulnerabilidades. La herramienta está escrita en Go y está diseñada para ser fácil de usar y ampliar.

## Características

- **Pruebas de Autenticación**: Verifica si los puntos de extremidad de la API requieren autenticación adecuada.
- **Validación de Métodos HTTP**: Asegura que los puntos de extremidad de la API admitan solo los métodos HTTP previstos.
- **Detección de Inyección SQL**: Identifica posibles vulnerabilidades de inyección SQL enviando cargas útiles y analizando las respuestas.
- **Informes Detallados**: Genera un informe detallado que detalla los resultados de cada prueba y proporciona una evaluación de seguridad general.
- **Pruebas Concurrentes**: Ejecuta pruebas de seguridad de forma simultánea para mejorar el rendimiento.
- **Configuración Personalizable**: Permite a los usuarios personalizar los puntos de extremidad, las credenciales de autenticación y las cargas útiles de inyección a través de un archivo de configuración.

## Instalación

Para instalar el API Security Scanner, siga los siguientes pasos:

1. **Requisitos Previos**: Asegúrese de tener Go instalado en su sistema. Puede descargarlo desde [aquí](https://golang.org/dl/).

2. **Clonar el Repositorio**:
   ```bash
   git clone https://github.com/elliotsecops/api-security-scanner.git
   cd api-security-scanner
   ```

3. **Compilar el Proyecto**:
   ```bash
   go build
   ```

4. **Ejecutar el Escáner**:
   ```bash
   ./api-security-scanner
   ```

## Configuración

El API Security Scanner utiliza un archivo de configuración YAML (`config.yaml`) para especificar los puntos de extremidad de la API, las credenciales de autenticación y las cargas útiles de inyección. Aquí hay un ejemplo de configuración:

```yaml
api_endpoints:
  - url: http://127.0.0.1:5000/basic-auth/admin/password
    method: GET
    body: ""
  - url: http://127.0.0.1:5000/post
    method: POST
    body: '{"key": "value"}'
auth:
  username: admin
  password: password
injection_payloads:
  - "' OR '1'='1"
  - "'; DROP TABLE users;--"
```

### Opciones de Configuración

- **api\_endpoints**: Una lista de puntos de extremidad de la API a probar. Cada punto de extremidad incluye:
  - **url**: La URL del punto de extremidad de la API.
  - **method**: El método HTTP a utilizar (por ejemplo, GET, POST).
  - **body**: El cuerpo de la solicitud (si corresponde).

- **auth**: Las credenciales de autenticación para los puntos de extremidad de la API.
  - **username**: El nombre de usuario para la autenticación básica.
  - **password**: La contraseña para la autenticación básica.

- **injection\_payloads**: Una lista de cargas útiles de inyección SQL a probar.

## Uso

Para ejecutar el API Security Scanner, utilice el siguiente comando:

```bash
./api-security-scanner
```

El escáner cargará la configuración desde `config.yaml`, ejecutará las pruebas de seguridad y generará un informe detallado.

### Salida Ejemplo

```bash
2024/09/25 00:37:28 Cargada la configuración: &{APIEndpoints:[{URL:http://127.0.0.1:5000/basic-auth/admin/password Method:GET Body:} {URL:http://127.0.0.1:5000/post Method:POST Body:{"key": "value"}}] Auth:{Username:admin Password:password} InjectionPayloads:[' OR '1'='1 '; DROP TABLE users;--]}
2024/09/25 00:37:28 Punto de extremidad: http://127.0.0.1:5000/basic-auth/admin/password, Método: GET
2024/09/25 00:37:28 Punto de extremidad: http://127.0.0.1:5000/post, Método: POST

Informe de Escaneo de Seguridad de la API
==================================

Punto de extremidad: http://127.0.0.1:5000/basic-auth/admin/password
Puntuación General: 100/100
Resultados de la Prueba:
- Prueba de Autenticación: APROBADO
  Detalles: Prueba de Autenticación Aprobada
- Prueba de Método HTTP: APROBADO
  Detalles: Prueba de Método HTTP Aprobada
- Prueba de Inyección: APROBADO
  Detalles: Prueba de Inyección Aprobada
Evaluación de Riesgos:
No se han detectado riesgos significativos.
------------------------

Punto de extremidad: http://127.0.0.1:5000/post
Puntuación General: 50/100
Resultados de la Prueba:
- Prueba de Autenticación: APROBADO
  Detalles: Prueba de Autenticación Aprobada
- Prueba de Método HTTP: APROBADO
  Detalles: Prueba de Método HTTP Aprobada
- Prueba de Inyección: FALLIDO
  Detalles: se ha detectado una posible inyección SQL con la carga útil: ' OR '1'='1
Evaluación de Riesgos:
- Las vulnerabilidades de inyección SQL representan un riesgo significativo de violación de datos.
------------------------

Evaluación de Seguridad General:
Puntuación de Seguridad Promedio: 75/100
Vulnerabilidades Críticas Detectadas: 1

Se han detectado riesgos de seguridad moderados. Aborde las vulnerabilidades identificadas de manera oportuna.
```

## Pruebas

El API Security Scanner incluye pruebas unitarias para garantizar la corrección de sus funciones. Para ejecutar las pruebas, utilice el siguiente comando:

```bash
go test ./...
```

## Contribuciones

Las contribuciones son bienvenidas. Si deseas contribuir al API Security Scanner, haz un fork y envía tu pull request.

## Licencia

El API Security Scanner está licenciado bajo la Licencia MIT. Consulte el archivo [LICENSE](LICENSE) para obtener más detalles.

---

# API Security Scanner

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Introduction

The API Security Scanner is a powerful tool designed to help developers and security professionals assess the security posture of their APIs. It performs a series of security tests, including authentication checks, HTTP method validation, and SQL injection detection, to identify potential vulnerabilities. The tool is written in Go and is designed to be easy to use and extend.

## Features

- **Authentication Testing**: Checks if the API endpoints require proper authentication.
- **HTTP Method Validation**: Ensures that the API endpoints support only the intended HTTP methods.
- **SQL Injection Detection**: Identifies potential SQL injection vulnerabilities by sending payloads and analyzing responses.
- **Detailed Reporting**: Generates a comprehensive report detailing the results of each test and providing an overall security assessment.
- **Concurrent Testing**: Runs security tests concurrently to improve performance.
- **Customizable Configuration**: Allows users to customize the endpoints, authentication credentials, and injection payloads via a configuration file.

## Installation

To install the API Security Scanner, follow these steps:

1. **Prerequisites**: Ensure you have Go installed on your system. You can download it from [here](https://golang.org/dl/).

2. **Clone the Repository**:
   ```bash
   git clone https://github.com/elliotsecops/api-security-scanner.git
   cd api-security-scanner
   ```

3. **Build the Project**:
   ```bash
   go build
   ```

4. **Run the Scanner**:
   ```bash
   ./api-security-scanner
   ```

## Configuration

The API Security Scanner uses a YAML configuration file (`config.yaml`) to specify the API endpoints, authentication credentials, and injection payloads. Here is an example configuration:

```yaml
api_endpoints:
  - url: http://127.0.0.1:5000/basic-auth/admin/password
    method: GET
    body: ""
  - url: http://127.0.0.1:5000/post
    method: POST
    body: '{"key": "value"}'
auth:
  username: admin
  password: password
injection_payloads:
  - "' OR '1'='1"
  - "'; DROP TABLE users;--"
```

### Configuration Options

- **api_endpoints**: A list of API endpoints to be tested. Each endpoint includes:
  - **url**: The URL of the API endpoint.
  - **method**: The HTTP method to be used (e.g., GET, POST).
  - **body**: The request body (if applicable).

- **auth**: Authentication credentials for the API endpoints.
  - **username**: The username for basic authentication.
  - **password**: The password for basic authentication.

- **injection_payloads**: A list of SQL injection payloads to be tested.

## Usage

To run the API Security Scanner, use the following command:

```bash
./api-security-scanner
```

The scanner will load the configuration from `config.yaml`, run the security tests, and generate a detailed report.

### Example Output

```bash
2024/09/25 00:37:28 Loaded configuration: &{APIEndpoints:[{URL:http://127.0.0.1:5000/basic-auth/admin/password Method:GET Body:} {URL:http://127.0.0.1:5000/post Method:POST Body:{"key": "value"}}] Auth:{Username:admin Password:password} InjectionPayloads:[' OR '1'='1 '; DROP TABLE users;--]}
2024/09/25 00:37:28 Endpoint: http://127.0.0.1:5000/basic-auth/admin/password, Method: GET
2024/09/25 00:37:28 Endpoint: http://127.0.0.1:5000/post, Method: POST

API Security Scan Detailed Report
==================================

Endpoint: http://127.0.0.1:5000/basic-auth/admin/password
Overall Score: 100/100
Test Results:
- Auth Test: PASSED
  Details: Auth Test Passed
- HTTP Method Test: PASSED
  Details: HTTP Method Test Passed
- Injection Test: PASSED
  Details: Injection Test Passed
Risk Assessment:
No significant risks detected.
------------------------

Endpoint: http://127.0.0.1:5000/post
Overall Score: 50/100
Test Results:
- Auth Test: PASSED
  Details: Auth Test Passed
- HTTP Method Test: PASSED
  Details: HTTP Method Test Passed
- Injection Test: FAILED
  Details: potential SQL injection detected with payload: ' OR '1'='1
Risk Assessment:
- SQL injection vulnerabilities pose a significant data breach risk.
------------------------

Overall Security Assessment:
Average Security Score: 75/100
Critical Vulnerabilities Detected: 1

Moderate security risks detected. Address identified vulnerabilities promptly.
```

## Testing

The API Security Scanner includes unit tests to ensure the correctness of its functions. To run the tests, use the following command:

```bash
go test ./...
```

## Contributing

Contributions are welcome! If you would like to contribute to the API Security Scanner, please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Make your changes and commit them (`git commit -am 'Add some feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Create a new Pull Request.

Please ensure that your code follows the existing coding style and includes appropriate tests.

## License

The API Security Scanner is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---
