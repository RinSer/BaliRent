import * as React from 'react';
//import Button from '@mui/material/Button';
import CssBaseline from '@mui/material/CssBaseline';
//import TextField from '@mui/material/TextField';
//import FormControlLabel from '@mui/material/FormControlLabel';
//import Checkbox from '@mui/material/Checkbox';
// import Link from '@mui/material/Link';
// import Grid from '@mui/material/Grid';
// import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import Copyright from './Copyright';
import Map from './Map';
//import * as api from '../api/calls';

const theme = createTheme();

export default class Property extends React.Component {
  constructor(props) {
    super(props);
    this.state = {signUp: false};
  }

  render() {
    return (
        <ThemeProvider theme={theme}>
          <Container component="main" maxWidth="md">
            <CssBaseline />
            <Map position={[-8.409518, 115.188919]} />
            <Copyright sx={{ mt: 8, mb: 4 }} />
          </Container>
        </ThemeProvider>
      );
  }

  handleSubmit = async (event) => {
    event.preventDefault();
    //const data = new FormData(event.currentTarget);
    alert("have sent!");
  };
}
