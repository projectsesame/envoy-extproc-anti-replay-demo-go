# envoy-extproc-anti-replay-demo-go

This repository contains a demo application written in Go that demonstrates the usage of Envoy's External Processor (ExtProc) filter to do `anti-replay` for HTTP request.

## Overview

The Envoy ExtProc filter allows you to offload request processing logic to an external process, enabling you to customize and extend Envoy's functionality. This demo application showcases how to implement an ExtProc filter in Go.

## Features

   + Integration with Envoy's External Processor filter
   + Customizable request processing logic
   + Demonstrates handling of HTTP requests in Go
   + Simple and easy-to-understand codebase

## Getting Started

To get started with the demo application, follow these steps:

  1. Clone the repository:
     ```
     git clone https://github.com/projectsesame/envoy-extproc-anti-replay-demo-go.git
     ```

  2. Build the Go application:
     ```
     go build .
     ```

  3. Run the application:
     ```
     ./envoy-extproc-anti-replay-demo-go anti-replay --log-stream --log-phases timespan 450
     ```
  4. Do request:
     ```shell
     curl --request POST \
     --url http://127.0.0.1:8080/post \
     --data '{
      "key": "value",
      "key2": "",
      "sign": "659876b30987883efdf178e69f062896",
      "nonce": "6062",
      "timestamp": "1712480920"
     }'

     ```

     Field Description:
       1. **sign**:  MD5(k1=v1&k2=v2&...kN=vN), the key-valure pairs order by key's ascending alphabetical. (**ignore the zero value pair**).

          eg:

          ```shell
          sign= MD5("key=value&nonce=6062&timestamp=1712480920") = 659876b30987883efdf178e69f062896
          ```
       2. **nonce**: the uuid that can only be used once within a timespan.

       3. **timestamp**: the current timestamp.

     PS:
       1. The use of the md5 here is only as a demo, in product,please use the signature algorithm likes **SHA256WithRSA**.

## Usage

The demo application listens for incoming GRPC requests on a specified port and performs custom processing logic. You can modify the processing logic in the application code according to your requirements.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.
License

This project is licensed under the Apache License Version 2.0. See the LICENSE file for details.
Acknowledgements

This demo application is based on the ExtProc filter demo(s) provided by [envoy-extproc-sdk-go](https://github.com/wrossmorrow/envoy-extproc-sdk-go). please visit it for more demos.

Special thanks to the community for their contributions and support.

## Contact

For any questions or inquiries, please feel free to reach out to us for any assistance or feedback.
