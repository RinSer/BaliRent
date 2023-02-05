import * as React from 'react';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import CssBaseline from '@mui/material/CssBaseline';
import TextField from '@mui/material/TextField';
//import FormControlLabel from '@mui/material/FormControlLabel';
//import Checkbox from '@mui/material/Checkbox';
import Link from '@mui/material/Link';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import Copyright from './Copyright';
import * as api from '../api/calls';

const theme = createTheme();

export default class Clock extends React.Component {
  constructor(props) {
    super(props);
    this.state = {signUp: false};
  }

  render() {
    return (
        <ThemeProvider theme={theme}>
          <Container component="main" maxWidth="xs">
            <CssBaseline />
            <Box
              sx={{
                marginTop: 8,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
              }}
            >
              <Avatar sx={{ m: 1, bgcolor: 'secondary.main' }}>
                <LockOutlinedIcon />
              </Avatar>
              <Typography component="h1" variant="h5">
                {this.state.signUp ? 'Sign Up' : 'Sign In'}
              </Typography>
              <Box component="form" onSubmit={this.handleSubmit} noValidate sx={{ mt: 1 }}>
              <TextField
                  margin="normal"
                  required
                  fullWidth
                  id="phone"
                  type="tel"
                  label="Phone Number"
                  name="phone"
                  autoComplete="phone"
                  autoFocus
                />
                {this.state.signUp ? <TextField
                  margin="normal"
                  required
                  fullWidth
                  id="email"
                  type="email"
                  label="Email Address"
                  name="email"
                  autoComplete="email"
                /> : null}
                <TextField
                  onChange={this.checkPasswordsMatch}
                  margin="normal"
                  required
                  fullWidth
                  name="password"
                  label="Password"
                  type="password"
                  id="password"
                  autoComplete="current-password"
                />
                {this.state.signUp ? <TextField
                  onChange={this.checkPasswordsMatch}
                  margin="normal"
                  required
                  fullWidth
                  name="repeat"
                  label="Repeat Password"
                  type="password"
                  id="repeat"
                /> : null}
                {/* <FormControlLabel
                  control={<Checkbox value="remember" color="primary" />}
                  label="Remember me"
                /> */}
                <Button
                  id="submit"
                  type="submit"
                  fullWidth
                  variant="contained"
                  sx={{ mt: 3, mb: 2 }}
                >
                  {this.state.signUp ? 'Sign Up' : 'Sign In'}
                </Button>
                <Grid container>
                  {/* <Grid item xs>
                    <Link href="#" variant="body2">
                      Forgot password?
                    </Link>
                  </Grid> */}
                  <Grid item>
                    <Link onClick={this.toggleForm} href="#" variant="body2">
                      {this.state.signUp
                        ? "Have an account? Sign In"
                        : "Don't have an account? Sign Up"}
                    </Link>
                  </Grid>
                </Grid>
              </Box>
            </Box>
            <Copyright sx={{ mt: 8, mb: 4 }} />
          </Container>
        </ThemeProvider>
      );
  }

  toggleForm = (event) => {
    event.preventDefault();
    this.setState({
      ...this.state,
      signUp: !this.state.signUp
    });
  };

  checkPasswordsMatch = (event) => {
    event.preventDefault();
    // const data = new FormData(event.currentTarget.form);
    // if (this.state.signUp) {
    //   const password = data.get('password');
    //   const repeat = data.get('repeat');
    //   if (password !== repeat) {
    //     document.getElementById("submit").setAttribute("disabled", "true");
    //   } else {
    //     document.getElementById("submit").setAttribute("disabled", "false");
    //   }
    // }
  };

  handleSubmit = async (event) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    if (this.state.signUp) {
      const phone = data.get('phone');
      const email = data.get('email');
      const password = data.get('password');
      try {
        const response = await api.register(phone, email, password);
        if (!response.id) {
          alert("Could not create new account!");
          return;
        }
      } catch(err) {
        alert("Could not create new account!");
        console.log(err);
        return;
      }
      this.setState({
        ...this.state,
        signUp: !this.state.signUp
      });
    } else {
      const phone = data.get('phone');
      const password = data.get('password');
      const response = await api.login(phone, password);
      localStorage.setItem('user', response);
      // TODO: remove !!!
      alert(response.token);
    }
  };
}
