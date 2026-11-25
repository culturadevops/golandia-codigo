/* AUTOMATICALLY GENERATED CODE DO NOT MODIFY */
/*   To generate run: "serverpod generate"    */

// ignore_for_file: implementation_imports
// ignore_for_file: library_private_types_in_public_api
// ignore_for_file: non_constant_identifier_names
// ignore_for_file: public_member_api_docs
// ignore_for_file: type_literal_in_constant_pattern
// ignore_for_file: use_super_parameters
// ignore_for_file: invalid_use_of_internal_member

// ignore_for_file: no_leading_underscores_for_library_prefixes
import 'package:serverpod/serverpod.dart' as _is;
import '../api/api_endpoint.dart' as _ib4l020b;

class Endpoints extends _is.EndpointDispatch {
  @override
  void initializeEndpoints(_is.Server server) {
    initializeOwnEndpoints(server);
  }

  @override
  void initializeOwnEndpoints(_is.Server server) {
    var endpoints = <String, _is.Endpoint>{
      'api': _ib4l020b.ApiEndpoint()
        ..initialize(
          server,
          'api',
          null,
        ),
    };
    connectors['api'] = _is.EndpointConnector(
      name: 'api',
      endpoint: endpoints['api']!,
      methodConnectors: {
        'tarea': _is.MethodConnector(
          name: 'tarea',
          params: {
            'id': _is.ParameterDescription(
              name: 'id',
              type: _is.getType<int>(),
              nullable: false,
            ),
          },
          call:
              (
                _is.Session session,
                Map<String, dynamic> params,
              ) async => (endpoints['api'] as _ib4l020b.ApiEndpoint).tarea(
                session,
                params['id'],
              ),
        ),
        'autores': _is.MethodConnector(
          name: 'autores',
          params: {
            'limite': _is.ParameterDescription(
              name: 'limite',
              type: _is.getType<int>(),
              nullable: false,
            ),
          },
          call:
              (
                _is.Session session,
                Map<String, dynamic> params,
              ) async => (endpoints['api'] as _ib4l020b.ApiEndpoint).autores(
                session,
                params['limite'],
              ),
        ),
        'autoresN1': _is.MethodConnector(
          name: 'autoresN1',
          params: {
            'limite': _is.ParameterDescription(
              name: 'limite',
              type: _is.getType<int>(),
              nullable: false,
            ),
          },
          call:
              (
                _is.Session session,
                Map<String, dynamic> params,
              ) async => (endpoints['api'] as _ib4l020b.ApiEndpoint).autoresN1(
                session,
                params['limite'],
              ),
        ),
        'crear': _is.MethodConnector(
          name: 'crear',
          params: {
            'titulo': _is.ParameterDescription(
              name: 'titulo',
              type: _is.getType<String>(),
              nullable: false,
            ),
            'autorId': _is.ParameterDescription(
              name: 'autorId',
              type: _is.getType<int>(),
              nullable: false,
            ),
          },
          call:
              (
                _is.Session session,
                Map<String, dynamic> params,
              ) async => (endpoints['api'] as _ib4l020b.ApiEndpoint).crear(
                session,
                params['titulo'],
                params['autorId'],
              ),
        ),
      },
    );
  }
}
