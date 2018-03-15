import { Namespace } from './Namespace';

export interface ServiceName {
  name: string;
}

export interface ServiceList {
  namespace: Namespace;
  services: ServiceName[];
}

export interface ServiceItem {
  servicename: string;
  namespace: string;
}
