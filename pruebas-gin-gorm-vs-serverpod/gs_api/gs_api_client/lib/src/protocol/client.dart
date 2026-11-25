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
import 'dart:async' as _ida;
import 'package:gs_api_client/src/protocol/modelos/autor.dart' as _i91hbr78;
import 'package:gs_api_client/src/protocol/modelos/tarea.dart' as _ix95efzw;
import 'package:http/http.dart' as _i85jenna;
import 'package:serverpod_client/serverpod_client.dart' as _isc;
import 'protocol.dart' as _il2as5qe;

/// La misma API que el servidor de Go (go/main.go): una tarea por id, los autores con sus tareas y crear una tarea.
/// {@category Endpoint}
class EndpointApi extends _isc.EndpointRef {
  EndpointApi(_isc.EndpointCaller caller) : super(caller);

  @override
  String get name => 'api';

  /// Una tarea por id.
  _ida.Future<_ix95efzw.Tarea?> tarea(int id) =>
      caller.callServerEndpoint<_ix95efzw.Tarea?>(
        'api',
        'tarea',
        {'id': id},
      );

  /// Los primeros [limite] autores con sus tareas, con include: lo que recomienda Serverpod.
  _ida.Future<List<_i91hbr78.Autor>> autores(int limite) =>
      caller.callServerEndpoint<List<_i91hbr78.Autor>>(
        'api',
        'autores',
        {'limite': limite},
      );

  /// Lo mismo, a mano: una consulta por autor (el N+1).
  _ida.Future<List<_i91hbr78.Autor>> autoresN1(int limite) =>
      caller.callServerEndpoint<List<_i91hbr78.Autor>>(
        'api',
        'autoresN1',
        {'limite': limite},
      );

  /// Crea una tarea y devuelve su id.
  _ida.Future<int> crear(
    String titulo,
    int autorId,
  ) => caller.callServerEndpoint<int>(
    'api',
    'crear',
    {
      'titulo': titulo,
      'autorId': autorId,
    },
  );
}

class Client extends _isc.ServerpodClientShared {
  Client(
    String host, {
    dynamic securityContext,
    Duration? streamingConnectionTimeout,
    Duration? connectionTimeout,
    Function(
      _isc.MethodCallContext,
      Object,
      StackTrace,
    )?
    onFailedCall,
    Function(_isc.MethodCallContext)? onSucceededCall,
    bool? disconnectStreamsOnLostInternetConnection,
    _i85jenna.Client? httpClientOverride,
  }) : super(
         host,
         _il2as5qe.Protocol(),
         securityContext: securityContext,
         streamingConnectionTimeout: streamingConnectionTimeout,
         connectionTimeout: connectionTimeout,
         onFailedCall: onFailedCall,
         onSucceededCall: onSucceededCall,
         disconnectStreamsOnLostInternetConnection:
             disconnectStreamsOnLostInternetConnection,
         httpClientOverride: httpClientOverride,
       ) {
    api = EndpointApi(this);
  }

  late final EndpointApi api;

  @override
  Map<String, _isc.EndpointRef> get endpointRefLookup => {'api': api};

  @override
  Map<String, _isc.ModuleEndpointCaller> get moduleLookup => {};
}
