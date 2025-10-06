import { app, port } from './presentation/app';

app.listen(port, () => {
  console.log('Server running on port:', port);
});