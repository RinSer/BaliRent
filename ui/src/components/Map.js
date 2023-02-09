import React from 'react'
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet'

export default class Property extends React.Component {
  constructor(props) {
    super(props);
    this.state = {markerPosition: props.position};
  }

  render() {
    return (
      <div>
      <MapContainer center={this.props.position} zoom={10} scrollWheelZoom={true} onClick={this.onClick}>
        <TileLayer
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <Marker position={this.state.markerPosition}>
          <Popup>
            A pretty CSS3 popup. <br /> Easily customizable.
          </Popup>
        </Marker>
      </MapContainer>
      </div>
    );
  }

  onClick = async (event) => {
    event.preventDefault();
    console.log(event);
  };
}


